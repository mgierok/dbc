---
name: write-commit-messages
description: Create Conventional Commits v1.0.0 commit messages from user intent, diff, or staged changes. This skill SHOULD be used when the user asks for a commit message, says "propose commit message", "write commit title", "name commit message", "commit name", "zaproponuj commit", "zaproponuj message do commita", "nazwa commita", "wymysl nazwe commita", or asks to commit without providing a message. The output MUST select type, optional scope, and breaking-change footer when needed. This skill MUST NOT be used for PR titles, release notes, or changelog generation unless explicitly requested.
---

# Write Commit Messages

## Critical Rules

- The response MUST contain exactly one of:
  - one clarifying question, or
  - ready commit message output.
- If enough context exists, the skill MUST NOT ask follow-up questions.
- The header MUST follow Conventional Commits format: `type(scope)!: description`.
- The description MUST be imperative and MUST NOT have a trailing period.
- The description SHOULD stay concise.
- If a change is breaking, the output MUST include `!` in the header and a `BREAKING CHANGE: ...` footer.
- If explicit user or repository conventions conflict with defaults, the skill MUST follow the explicit convention.

## Workflow

1. The skill MUST gather context in this order: explicit user description, provided diff, staged or changed file names, repository conventions.
2. If intent is still unclear, the skill MUST ask one short clarifying question focused on missing information.
3. The skill MUST choose commit type based on behavioral intent, not file extension alone.
4. The skill SHOULD add scope only when it improves clarity, and scope SHOULD be short and specific.
5. The skill MUST detect breaking changes and MUST apply both required markers.
6. The skill MUST draft the subject first, then optional body and optional footers.
7. The skill MUST run the validation checklist before returning output.

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

- Header MUST match `type(scope)!: description` with optional `scope` and optional `!`.
- Type MUST reflect change intent.
- Scope MUST be omitted when it is unclear or too broad.
- Subject MUST NOT have a trailing period.
- Breaking changes MUST include both `!` and `BREAKING CHANGE:` footer.
- Body and footers SHOULD be included only when they add useful context.

## Troubleshooting

- Ambiguous mixed changes:
The skill MUST ask whether the user wants one commit or multiple commits split by concern.

- Missing context:
The skill MUST ask for a short summary of what changed and why, or request diff or staged files.

- Multiple valid types:
The skill SHOULD prefer the type that best represents externally visible impact.

- Non-Conventional repo style:
If repository convention is explicit, the skill MUST follow repository convention and SHOULD note deviation only if asked.

## Output Rules

- When information is sufficient, the response MUST return only commit message text in a fenced `text` block.
- For multiple commits, the response MUST return numbered commit messages, each in its own fenced `text` block.
- When information is insufficient, the response MUST return one concise clarifying question.

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
