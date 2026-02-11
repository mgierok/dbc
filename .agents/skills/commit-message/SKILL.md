---
name: commit-message
description: Create Conventional Commits v1.0.0 commit messages from user intent, diff, or staged changes. Use when the user asks for a commit message, says "propose commit message", "write commit title", "name commit message", "commit name", "zaproponuj commit", "zaproponuj message do commita", "nazwa commita", "wymysl nazwe commita", or asks to commit without providing a message. Select type, optional scope, and breaking-change footer when needed. Do not use for PR titles, release notes, or changelog generation unless explicitly requested.
---

# Write Commit Messages

## Critical Rules

- Return either one clarifying question or ready commit message output, never both in the same response.
- If enough context exists, do not ask follow-up questions.
- Follow Conventional Commits header format: `type(scope)!: description`.
- Keep description imperative, concise, and without trailing period.
- If change is breaking, add `!` in header and `BREAKING CHANGE: ...` footer.
- Follow explicit user or repository conventions when they conflict with defaults.

## Workflow

1. Gather context in this order: explicit user description, provided diff, staged or changed file names, repository conventions.
2. If intent is still unclear, ask one short clarifying question focused on missing information.
3. Choose commit type based on behavioral intent, not file extension alone.
4. Decide scope only when it adds clarity; keep it short and specific.
5. Detect breaking change and apply both required markers.
6. Draft subject first, then optional body and optional footers.
7. Run the validation checklist before returning output.

## Type Selection Guide

- `feat`: introduces new user-facing capability.
- `fix`: corrects incorrect behavior or bug.
- `docs`: documentation-only changes.
- `style`: formatting-only changes, no runtime behavior changes.
- `refactor`: code restructuring without behavior change.
- `perf`: measurable performance improvement.
- `test`: adds or updates tests only.
- `build`: build/dependency/tooling changes affecting build process.
- `ci`: CI or CD pipeline or workflow config changes.
- `chore`: maintenance changes that do not fit other types.

## Validation Checklist

- Header matches `type(scope)!: description` with optional `scope` and optional `!`.
- Type reflects intent of change.
- Scope is omitted if unclear or too broad.
- Subject has no trailing period.
- Breaking changes include both `!` and `BREAKING CHANGE:` footer.
- Body and footers are included only when they add useful context.

## Troubleshooting

- Ambiguous mixed changes:
Ask whether user wants one commit or multiple commits split by concern.

- Missing context:
Ask for a short summary of what changed and why, or request diff or staged files.

- Multiple valid types:
Prefer the type that best represents externally visible impact.

- Non-Conventional repo style:
If repository convention is explicit, follow repository convention and note deviation only if asked.

## Output Rules

- When information is sufficient, return only commit message text in a fenced `text` block.
- For multiple commits, return numbered commit messages, each in its own fenced `text` block.
- When information is insufficient, return one concise clarifying question.

## Examples

```text
feat(parser): add support for quoted tokens
```

```text
fix(api)!: reject unknown fields in request payloads

BREAKING CHANGE: clients must remove unknown fields before sending requests
```

```text
docs(readme): add local setup and troubleshooting steps
```
