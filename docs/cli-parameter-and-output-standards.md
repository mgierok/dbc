# CLI Parameter and Output Standards

## 1. Purpose

Define a reusable, implementation-agnostic standard for:

- command-line parameter design,
- input validation and argument error handling,
- human-readable and machine-readable output formats.

The standard is based on proven conventions used in widely adopted CLI tools (for example: `git`, `docker`, `kubectl`, `curl`, `jq`).

## 2. Baseline for Further Analysis (Anonymized)

Current input audit (treated as baseline, without product context):

- one value-carrying option exists with both short and long alias,
- option value is mandatory and must not be empty,
- unsupported options are rejected immediately,
- duplicate usage of a logically identical option is rejected,
- parse errors return actionable usage hints.

This baseline is valid but minimal. The sections below define a scalable standard for broader CLI evolution.

## 3. Command Model

Use this canonical shape:

```text
tool [global-options] <command> [command-options] [arguments]
```

Rules:

- global options are available before and after command name only if parser supports both forms consistently,
- command options apply only to a single command,
- positional arguments are explicit and ordered,
- hidden implicit arguments are disallowed.

## 4. Parameter Design Standard

### 4.1 Naming

- long options use `kebab-case` (example: `--config-path`),
- short options are single-letter aliases for high-frequency operations (example: `-c`),
- booleans default to `false` and switch to `true` when present,
- value options use one of two allowed forms:
  - `--option <value>`
  - `--option=<value>`

### 4.2 Aliases and Compatibility

- at most one short alias per long option,
- alias pairs must be documented in help output,
- removing an alias requires deprecation period and warning message.

### 4.3 Repetition and Conflicts

- repeated single-value option: reject as usage error unless "last wins" is explicitly documented,
- repeated multi-value option: append values in order,
- mutually exclusive options: reject with explicit conflict message,
- required-together options: reject when partial set is provided.

### 4.4 Type and Range Validation

Validate before runtime action starts:

- non-empty string,
- path or URI syntax (when relevant),
- integer/float range,
- enum allowlist,
- file existence/permission checks (when required by command semantics).

## 5. Input Error Standard

### 5.1 Behavior

- fail fast on first invalid argument,
- do not run side effects when argument validation fails,
- print errors to `stderr`,
- return non-zero exit code.

### 5.2 Error Message Format

Use this text structure:

```text
Error: <problem summary>.
Hint: <how to fix>.
Usage: <short canonical usage line>.
```

Message quality rules:

- include offending token when possible,
- include one corrective action,
- avoid internal stack traces in default mode.

## 6. Output Standard for Humans (`text`)

### 6.1 Streams

- primary command result -> `stdout`,
- diagnostics, warnings, and errors -> `stderr`.

### 6.2 Formatting

- keep one logical message per line,
- stable label vocabulary (`INFO`, `WARN`, `ERROR`) when prefixes are used,
- avoid decorative noise and unstable wording in automation-sensitive commands.

### 6.3 Verbosity Controls

- `--quiet`: only essential result or fatal error,
- default: concise operational summary,
- `--verbose`: additional diagnostic context.

## 7. Machine-Readable Output Standard

Support explicit output selection:

```text
--output text|json|yaml
```

Rules:

- never mix human commentary with `json`/`yaml` on `stdout`,
- schema fields remain backward-compatible across patch/minor releases,
- error payload for machine modes should include:
  - `code`
  - `message`
  - `hint` (optional)
  - `details` (optional structured object)

## 8. Exit Codes

Recommended baseline:

- `0` success,
- `1` runtime/operation failure,
- `2` invalid usage or argument validation failure,
- `3` configuration or permission failure,
- `4` dependency/unavailable external system,
- `130` interrupted by user (`SIGINT`).

If command-specific codes are added, they must be documented in help and reference docs.

## 9. Help and Discoverability

Every command should provide:

- short description,
- canonical usage line,
- option table with short/long aliases and value expectations,
- 2-4 practical examples,
- reference to `--help` at every parse error.

Help output should be deterministic and suitable for copy-paste execution.

## 10. Standardized CLI Audit Checklist

Use this checklist for recurring analysis:

1. Are all options named consistently (`kebab-case`, clear semantics)?
2. Are short aliases present only for frequent options?
3. Are conflicts and required-together rules enforced?
4. Are validation errors actionable and deterministic?
5. Are `stdout` and `stderr` strictly separated?
6. Is machine output clean and schema-stable?
7. Are exit codes mapped and documented?
8. Does `--help` include usable examples?
