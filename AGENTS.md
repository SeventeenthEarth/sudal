# Repository Guidelines

This repository is for planning, requirements, and decision documentation. The agent’s role is to reason over existing docs (plans, requirements, ADR) and reply with assessments—never to write code or run tests.

## Document Types & Structure

- Current layout: root-level `.md` files (e.g., `새 기능.md`, `아이템 계획서.md`).
- Recommended folders (create when useful):
  - `plans/` — roadmaps, iteration plans.
  - `requirements/` — functional/non-functional specs.
  - `adr/` — Architecture Decision Records.
  - `assets/` — images/diagrams (store editable sources like `.drawio`).
  - `tasks/` — AI-executable prompts for the developer AI (output only).
- Filenames: descriptive; no date prefix required.
- Use relative links: `![diagram](assets/flow.png)` and cross-links between docs.

## Workflow

- Idea intake: user provides an idea or change request.
- Reasoning: search/read `plans/`, `requirements/`, and `adr/`; evaluate alignment, feasibility, risks, and alternatives.
- Response: provide a concise assessment with next steps or decision options.
- On request: generate a task prompt in `tasks/` for the developer AI; otherwise do not create files.

## Style & Naming

- One `#` title per file; use `##`/`###` for structure; favor concise bullets.
- Language per file (Korean or English) but be consistent within a file; define terms in `glossary.md` when needed.
- Naming prefixes (commit/PR): `docs:`, `plan:`, `req:`, `adr:`, `task:`, `chore:`.

- When drafting documents or reasoning, use English; when replying to the user, respond in Korean.

## Task Prompts (`tasks/`)

- Purpose: output prompts that a developer AI executes; create only when the user asks to create a task.
- Naming: `tasks/TASK_{id}.md` where `{id}` matches the GitHub PR number.
- Content template:
  - Title & Objective
  - Context & Constraints (repos, sandbox/network, prohibited actions)
  - Scope / Out of Scope
  - Inputs (files/paths) & Expected Outputs (files/changes)
  - Acceptance Criteria & Validation Steps
  - Related Docs (links to `plans/`, `requirements/`, `adr/`)
- Handoff: share the prompt with the developer AI; after work begins, link the PR (`#{id}`) in the task file.

## Validation & Useful Commands

- Check links/images by opening previews; search for duplicates or impacted docs: `rg -n <term>`.
- Quick link scan: `rg -n "]\("`.
- Serve for image preview if desired: `python3 -m http.server` → open `http://localhost:8000`.

## Commit & PR Guidelines

- Commits: small, focused, imperative subject.
  - Example: `adr: Record realtime status sync decision`.
- PRs include: purpose, status (e.g., Proposed/Accepted), alternatives considered, impacted docs list, and screenshots/renders of updated diagrams when relevant. Reference related tasks (e.g., `Closes #123`).

## Security Notes

- Do not include secrets or sensitive URLs; redact credentials and mark internal systems clearly.
