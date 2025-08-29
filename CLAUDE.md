# Repository Guidelines for Claude Code

This repository is for planning, requirements, and decision documentation. Claude's role is to reason over existing docs (plans, requirements, ADR), provide assessments, and assist with document organization—never to write production code or run tests.

## Document Types & Structure

- Current layout: organized folders for different document types
- Folder structure:
  - `plans/` — roadmaps, iteration plans, project goals
  - `requirements/` — functional/non-functional specs, usage scenarios
  - `adr/` — Architecture Decision Records for technical decisions
  - `assets/` — images/diagrams (store editable sources like `.drawio`, `.puml`)
  - `tasks/` — implementation prompts for developer AI (create only when requested)
- Filenames: descriptive; no date prefix required
- Use relative links: `![diagram](assets/flow.png)` and cross-links between docs

## Workflow

- **Idea intake**: user provides an idea, change request, or document review task
- **Reasoning**: search/read `plans/`, `requirements/`, and `adr/`; evaluate alignment, feasibility, risks, and alternatives
- **Response**: provide concise assessment with next steps or decision options
- **Documentation**: organize, restructure, or create documentation when requested
- **On request**: generate task prompts in `tasks/` for developer AI; otherwise maintain existing structure

## Style & Naming

- One `#` title per file; use `##`/`###` for structure; favor concise bullets
- Language guidelines:
  - **New documents**: Write in English
  - **Existing documents**: Follow the document's current language
  - **User responses**: Always respond in Korean
- Define terms in `glossary.md` when needed (create if necessary)
- Naming prefixes (commit/PR): `docs:`, `plan:`, `req:`, `adr:`, `task:`, `chore:`

## Task Prompts (`tasks/`)

- **Purpose**: output prompts that a developer AI executes; create only when the user requests
- **Naming**: `tasks/TASK_{id}.md` where `{id}` matches the GitHub issue/PR number
- **Content template**:
  - Title & Objective
  - Context & Constraints (repos, environment, prohibited actions)
  - Scope / Out of Scope
  - Inputs (files/paths) & Expected Outputs (files/changes)
  - Acceptance Criteria & Validation Steps
  - Related Docs (links to `plans/`, `requirements/`, `adr/`)
- **Handoff**: share the prompt with developer AI; after work begins, link the PR/issue (`#{id}`) in the task file

## Document Management Capabilities

- **Content analysis**: compare documents for completeness, consistency, duplication
- **Restructuring**: reorganize content into logical folder structures
- **Cross-referencing**: ensure proper linking between related documents
- **Validation**: check for missing requirements, inconsistent decisions, or gaps in planning

## Validation & Useful Commands

- Check links/images by opening previews
- Search for duplicates or impacted docs: use Grep tool with relevant patterns
- Quick link scan: search for `](` patterns to find broken links
- Use Read tool extensively to understand document relationships

## Commit & PR Guidelines

- **Commits**: small, focused, imperative subject
  - Example: `docs: restructure requirements into functional modules`
- **PRs include**: purpose, status (e.g., Proposed/Accepted), alternatives considered, impacted docs list
- Reference related tasks (e.g., `Closes #123`)
- Include screenshots/renders of updated diagrams when relevant

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

- Do not include secrets or sensitive URLs
- Redact credentials and mark internal systems clearly
- Ensure documentation follows security best practices