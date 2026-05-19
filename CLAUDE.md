# Building with the Claude API — Go Edition

Go implementations of the coding examples from Anthropic's [Claude with the Anthropic API](https://anthropic.skilljar.com/claude-with-the-anthropic-api) course.

## Project Structure

Each course module gets its own package under a top-level directory matching the module name. Example implementations live as runnable `main` packages inside those directories; shared utilities (if any) live in `internal/`.

```
building-with-claude-api/
├── main.go                  # entry point / scratch
├── go.mod
├── 01-api-basics/           # Accessing Claude with the API
├── 02-prompt-evaluation/    # Prompt Evaluation
├── 03-prompt-engineering/   # Prompt Engineering Techniques
├── 04-tool-use/             # Tool Use with Claude
├── 05-rag/                  # RAG and Agentic Search
├── 06-claude-features/      # Extended thinking, vision, caching, Files API, etc.
├── 07-mcp/                  # Model Context Protocol
├── 08-agents/               # Agents and Workflows
└── internal/                # Shared helpers
```

## Module Coverage

| # | Module | Key topics |
|---|--------|-----------|
| 1 | API Basics | Basic requests, multi-turn conversations, system prompts, temperature, streaming, structured output |
| 2 | Prompt Evaluation | Eval workflows, test dataset generation, model-graded and code-graded evals |
| 3 | Prompt Engineering | Clarity, specificity, XML tags, few-shot examples |
| 4 | Tool Use | Tool schemas, message block handling, multi-turn tool loops, multiple tools |
| 5 | RAG | Text chunking, embeddings, BM25 lexical search, multi-index pipelines |
| 6 | Claude Features | Extended thinking, image/PDF input, citations, prompt caching, code execution, Files API |
| 7 | MCP | MCP clients and servers, tool and resource definitions |
| 8 | Agents & Workflows | Parallelization, chaining, routing, agent design patterns |

## Tech Stack

- **Language:** Go 1.26+
- **SDK:** `github.com/anthropic-ai/anthropic-sdk-go` (official Anthropic Go SDK)
- **Model defaults:** `claude-sonnet-4-6` for general use; `claude-opus-4-7` where extended thinking is exercised

## Running Examples

Each subdirectory is a standalone runnable program:

```bash
go run ./01-api-basics/basic-request/
```

Environment variables required:

```
ANTHROPIC_API_KEY=sk-ant-...
```

## Conventions

- One example per subdirectory, named after what it demonstrates (e.g., `streaming`, `tool-loop`, `prompt-caching`).
- No global state; each example is self-contained.
- Prompt caching headers included wherever the SDK supports it (cost and latency matter).
- Errors are handled explicitly — no `panic` outside of `main` init paths.
