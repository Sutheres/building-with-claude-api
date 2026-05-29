package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

const model = anthropic.ModelClaudeHaiku4_5

// testCase is one row of the evaluation dataset.
type testCase struct {
	Task             string `json:"task"`
	Format           string `json:"format"`
	SolutionCriteria string `json:"solution_criteria"`
}

// modelGrade is the structured judgement returned by the grader model.
type modelGrade struct {
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
	Reasoning  string   `json:"reasoning"`
	Score      int      `json:"score"`
}

// result is the graded outcome of running one test case.
type result struct {
	TestCase  testCase `json:"test_case"`
	Output    string   `json:"output"`
	Score     float64  `json:"score"`
	Reasoning string   `json:"reasoning"`
}

func main() {
	datasetPath := flag.String("dataset", "02-prompt-evaluation/dataset.json", "path to the eval dataset")
	flag.Parse()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()

	data, err := os.ReadFile(*datasetPath)
	if err != nil {
		log.Fatal(err)
	}
	var dataset []testCase
	if err := json.Unmarshal(data, &dataset); err != nil {
		log.Fatal(err)
	}

	results, avg, err := runEval(context.Background(), client, dataset)
	if err != nil {
		log.Fatal(err)
	}

	for i, r := range results {
		fmt.Printf("--- case %d [%s] score %.1f ---\n", i+1, r.TestCase.Format, r.Score)
		fmt.Printf("task: %s\n", r.TestCase.Task)
		fmt.Printf("reasoning: %s\n\n", r.Reasoning)
	}
	fmt.Printf("Average score: %.2f\n", avg)
}

// runEval grades every case in the dataset and returns the per-case results
// plus the mean score.
func runEval(ctx context.Context, client anthropic.Client, dataset []testCase) ([]result, float64, error) {
	results := make([]result, 0, len(dataset))
	var total float64
	for _, tc := range dataset {
		r, err := runTestCase(ctx, client, tc)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, r)
		total += r.Score
	}
	if len(results) == 0 {
		return results, 0, nil
	}
	return results, total / float64(len(results)), nil
}

// runTestCase produces an output for the task, then blends two independent
// grades: a model-graded judgement (does it actually solve the task?) and a
// code-graded syntax check (is the output well-formed?).
func runTestCase(ctx context.Context, client anthropic.Client, tc testCase) (result, error) {
	output, err := runPrompt(ctx, client, tc)
	if err != nil {
		return result{}, err
	}

	grade, err := gradeByModel(ctx, client, tc, output)
	if err != nil {
		return result{}, err
	}

	syntaxScore := gradeSyntax(output, tc.Format)
	score := (float64(grade.Score) + float64(syntaxScore)) / 2

	return result{
		TestCase:  tc,
		Output:    output,
		Score:     score,
		Reasoning: grade.Reasoning,
	}, nil
}

// runPrompt is the prompt under evaluation: solve the task, emitting only
// code. We prefill a fence and stop at the closing fence so the output is
// the bare solution with no commentary.
func runPrompt(ctx context.Context, client anthropic.Client, tc testCase) (string, error) {
	prompt := fmt.Sprintf(`Please solve the following task:

%s

* Respond only with Go, JSON, or a plain Regex.
* Do not add any comments or commentary or explanation.`, tc.Task)

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:         model,
		MaxTokens:     1000,
		StopSequences: []string{"```"},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("```")),
		},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(collectText(resp)), nil
}

const gradePrompt = `You are an expert AWS code reviewer. Evaluate the following AI-generated solution.

Original Task:
<task>
%s
</task>

Solution to Evaluate:
<solution>
%s
</solution>

Criteria you should use to evaluate the solution:
<criteria>
%s
</criteria>

Provide your evaluation as a JSON object with these fields, in this order:
- "strengths": an array of 1-3 key strengths
- "weaknesses": an array of 1-3 key areas for improvement
- "reasoning": a concise explanation of your overall assessment
- "score": a number between 1 and 10

Respond only with JSON. Keep it concise and direct.`

// gradeByModel asks a model to judge the solution against the task criteria,
// returning a structured grade. This is the "model-graded" half of the eval.
func gradeByModel(ctx context.Context, client anthropic.Client, tc testCase, output string) (modelGrade, error) {
	prompt := fmt.Sprintf(gradePrompt, tc.Task, output, tc.SolutionCriteria)

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:         model,
		MaxTokens:     1000,
		StopSequences: []string{"```"},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("```json")),
		},
	})
	if err != nil {
		return modelGrade{}, err
	}

	var grade modelGrade
	if err := json.Unmarshal([]byte(strings.TrimSpace(collectText(resp))), &grade); err != nil {
		return modelGrade{}, fmt.Errorf("grader did not return valid JSON: %w", err)
	}
	return grade, nil
}

// gradeSyntax is the "code-graded" half: a deterministic check that the
// output is well-formed for its declared format. Returns 10 for valid, 0 for
// invalid. Unlike the Python course (which validates Python via ast.parse),
// the Go edition validates Go source with go/parser.
func gradeSyntax(output, format string) int {
	output = strings.TrimSpace(output)
	switch format {
	case "json":
		if json.Valid([]byte(output)) {
			return 10
		}
	case "go":
		if validGo(output) {
			return 10
		}
	case "regex":
		if _, err := regexp.Compile(output); err == nil {
			return 10
		}
	}
	return 0
}

// validGo reports whether src parses as Go. We try it as a complete file
// first, then as a package-level snippet (e.g. a bare function declaration).
func validGo(src string) bool {
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "", src, parser.SkipObjectResolution); err == nil {
		return true
	}
	_, err := parser.ParseFile(fset, "", "package p\n"+src, parser.SkipObjectResolution)
	return err == nil
}

func collectText(resp *anthropic.Message) string {
	var out string
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			out += text.Text
		}
	}
	return out
}
