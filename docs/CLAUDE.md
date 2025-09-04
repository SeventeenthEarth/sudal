# Repository Guidelines for Claude Code

This repository is for planning, requirements, and decision documentation. Claude's role is to reason over existing docs (plans, requirements, ADR), provide assessments, and assist with document organization—never to write production code or run tests.

## Document Types & Structure

- Docs live under `docs/` with these folders:
  - `plan/` — roadmaps, iteration plans
  - `requirement/` — functional/non-functional specs
  - `adr/` — Architecture Decision Records
  - `assets/` — images/diagrams (store editable sources like `.drawio`)
  - `tasks/` — AI-executable prompts for the developer AI (create only on request)
- Keep meta-guides (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`) at `docs/` root
- Filenames: descriptive; no date prefix required
- Use relative links: `![diagram](assets/flow.png)` and cross-links between docs

## Codebase Pointers (for cross-references)

- API contracts:
  - `api/openapi.yaml`: Authoritative REST spec. Server code is generated into `internal/infrastructure/openapi/*` via `ogen`; Swagger served under `/docs`.
  - `proto/*`: Authoritative gRPC/Connect contracts. Generated code lives in `gen/go` (protobuf, gRPC, Connect) and `gen/openapi` (OpenAPI v2).
- Application wiring:
  - `internal/infrastructure/server/*`: Unified HTTP server, route registry (REST + Connect-go), middleware chains.
  - `internal/infrastructure/di/wire.go`: Wire provider sets and initializers for services/handlers.
  - `internal/infrastructure/apispec/paths.go`: Protected procedure list for selective authentication.
- Services and infra modules (cross-cutting):
  - `internal/service/config`, `logger`, `postgres`, `sql/postgres`, `redis`, `cache`, `firebaseauth`, `authutil`, `health`.
- Data and feature domain:
  - `internal/feature/<name>/{domain,data,application,protocol}` keeps feature code cohesive.
- Config & migrations:
  - `.env.template` (local dev), `configs/config.yaml` (optional file input), `db/migrations/*`.

## Workflow

- Idea intake: user provides an idea or change request
- Reasoning: search/read `plans/`, `requirements/`, and `adr/`; evaluate alignment, feasibility, risks, and alternatives
- Response: provide a concise assessment with next steps or decision options
- On request: generate a task prompt in `tasks/` for the developer AI; otherwise do not create files

## Style & Naming

- One `#` title per file; use `##`/`###` for structure; favor concise bullets
- Language per file (Korean or English) but be consistent within a file; define terms in `glossary.md` when needed
- Naming prefixes (commit/PR): `docs:`, `plan:`, `req:`, `adr:`, `task:`, `chore:`

- When drafting documents or reasoning, use English; when replying to the user, respond in Korean

## Task Prompts (`tasks/`)

- Purpose: output prompts that a developer AI executes; create only when the user asks to create a task
- Naming: `tasks/TASK_{id}.md` where `{id}` matches the GitHub PR number
- Content template:
  - Title & Objective
  - Context & Constraints (repos, sandbox/network, prohibited actions)
  - Scope / Out of Scope
  - Inputs (files/paths) & Expected Outputs (files/changes)
  - Acceptance Criteria & Validation Steps
  - Related Docs (links to `plans/`, `requirements/`, `adr/`)
- Handoff: share the prompt with the developer AI; after work begins, link the PR (`#{id}`) in the task file


## Validation & Useful Commands

- Check links/images by opening previews; search for duplicates or impacted docs: `rg -n <term>`
- Quick link scan: `rg -n "]\("`
- Serve for image preview if desired: `python3 -m http.server` → open `http://localhost:8000`
- Cross-check API sources of truth (do not reverse‑engineer handlers):
  - REST: `api/openapi.yaml` → `internal/infrastructure/openapi/*` (ogen output)
  - gRPC/Connect: `proto/*` → `gen/go/*` and `gen/openapi/*`
- Reference build/test/generation commands (do not execute here):
  - `make generate-buf`, `make generate-ogen`, `make generate-wire`, `make generate`
  - `make test`, `make test.unit`, `make test.int`, `make test.e2e`
- Docs subtree management: `make push-docs`, `make pull-docs`

## Commit & PR Guidelines

- Commits: small, focused, imperative subject
  - Example: `adr: Record realtime status sync decision`
- PRs include: purpose, status (e.g., Proposed/Accepted), alternatives considered, impacted docs list, and screenshots/renders of updated diagrams when relevant. Reference related tasks (e.g., `Closes #123`)


## Claude-Specific Notes

- Use TodoWrite tool to track multi-step document management tasks
- Leverage parallel tool calls (Read, Grep, Glob) for efficient document analysis  
- Always read existing documents before making structural changes
- For large files, use offset/limit parameters to read specific portions
- Use Plan mode for complex changes - present plan with ExitPlanMode before execution
- Maintain document history and rationale for organizational decisions
- Focus on document quality, consistency, and maintainability

## Document Migration & Cleanup

- **Content verification**: Compare original and reorganized files for completeness
- **Duplication handling**: Identify and consolidate duplicate content across files
- **Safe deletion**: Verify all important content is preserved before removing original files
- **Cross-reference updating**: Ensure internal links remain valid after restructuring

## ADR Guidelines

- **Filename format**: `{topic}-{decision}.md` (e.g., `realtime-architecture.md`)
- **Required sections**: Status, Context, Decision, Consequences
- **Status labels**: Proposed, Accepted, Deprecated, Superseded
- **Numbering**: Use descriptive names rather than sequential numbers

## Security Notes

- Do not include secrets or sensitive URLs; redact credentials and mark internal systems clearly
- Follow repository secrets policy: use `.env` locally; never commit files under `secrets/` or credential keys