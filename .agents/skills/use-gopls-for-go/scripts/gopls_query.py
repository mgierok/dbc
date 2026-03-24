#!/usr/bin/env python3

import argparse
import json
import pathlib
import re
import subprocess
import sys
import urllib.parse


LOCATION_RE = re.compile(
    r"^(?P<path>.+?):(?P<line>\d+):(?P<column>\d+)(?:-(?:(?P<end_line>\d+):)?(?P<end_column>\d+))?$"
)
SYMBOL_RE = re.compile(
    r"^(?P<location>.+?:\d+:\d+(?:-(?:(?:\d+:)?\d+))?) (?P<name>.+?) (?P<kind>\S+)$"
)
CHECK_RE = re.compile(
    r"^(?P<location>.+?:\d+:\d+(?:-(?:(?:\d+:)?\d+))?): (?P<message>.+)$"
)
CALL_EDGE_RE = re.compile(
    r"^(?P<section>caller|callee)\[(?P<index>\d+)\]: ranges "
    r"(?P<selection>\d+:\d+(?:-(?:(?:\d+:)?\d+))?) in "
    r"(?P<selection_path>.+?) from/to "
    r"(?P<symbol_kind>\w+) (?P<symbol_name>.+?) in "
    r"(?P<symbol_location>.+)$"
)
CALL_IDENTIFIER_RE = re.compile(
    r"^identifier: (?P<symbol_kind>\w+) (?P<symbol_name>.+?) in (?P<symbol_location>.+)$"
)
OUTLINE_RE = re.compile(
    r"^(?P<indent>\t*)(?P<name>.+?) (?P<kind>\w+) (?P<range>\d+:\d+-\d+:\d+)$"
)
LINE_SCOPE_RE = re.compile(r"^\d+$")
LINE_RANGE_SCOPE_RE = re.compile(r"^(?P<start>\d+),(?P<end>\d+)$")
WORKSPACE_MARKERS = ("go.work", "go.mod")
CURSOR_MARKER = "<|>"


class UserError(Exception):
    def __init__(self, code, message, details=None):
        super().__init__(message)
        self.code = code
        self.message = message
        self.details = details or {}


def build_parser():
    parser = argparse.ArgumentParser(
        description="Run read-only or diff-only gopls queries and normalize the output as JSON."
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    search = subparsers.add_parser("search")
    search.add_argument("--workspace", required=True)
    search.add_argument("--query", required=True)

    outline = subparsers.add_parser("outline")
    outline.add_argument("--file", required=True)

    diagnostics = subparsers.add_parser("diagnostics")
    diagnostics.add_argument("--file", required=True)

    for name in (
        "locate",
        "definition",
        "references",
        "implementation",
        "call_hierarchy",
        "signature",
        "highlight",
        "rename_validate",
    ):
        command = subparsers.add_parser(name)
        add_scope_args(command)

    rename_preview = subparsers.add_parser("rename_preview")
    add_scope_args(rename_preview)
    rename_preview.add_argument("--name", required=True)

    return parser


def add_scope_args(parser):
    parser.add_argument("--file", required=True)
    parser.add_argument("--scope", required=True)
    parser.add_argument("--find")


def main(argv=None, runner=None):
    try:
        args = build_parser().parse_args(argv)
    except SystemExit as exc:
        return int(exc.code)

    run = runner or run_gopls
    try:
        payload = execute(args, run)
    except UserError as exc:
        emit_error(exc)
        return 1
    except Exception as exc:  # pragma: no cover - defensive fallback
        emit_error(UserError("internal_error", f"Unexpected failure: {exc}"))
        return 1

    json.dump(payload, sys.stdout, indent=2)
    sys.stdout.write("\n")
    return 0


def execute(args, runner):
    if args.command == "search":
        workspace = normalize_workspace(args.workspace)
        output = invoke_runner(
            runner, ["gopls", "workspace_symbol", args.query], str(workspace)
        )
        symbols = parse_workspace_symbols(output)
        return make_payload(
            "search",
            {"workspace": str(workspace), "query": args.query},
            {"symbols": symbols},
            str(workspace),
            result_count=len(symbols),
        )

    file_path = normalize_file(args.file)
    workspace = infer_workspace(file_path)

    if args.command == "outline":
        output = invoke_runner(
            runner, ["gopls", "symbols", str(file_path)], str(workspace)
        )
        symbols = parse_outline_symbols(output, file_path)
        return make_payload(
            "outline",
            {"file": str(file_path), "workspace": str(workspace)},
            {"symbols": symbols},
            str(workspace),
            result_count=len(symbols),
        )

    if args.command == "diagnostics":
        output = invoke_runner(
            runner, ["gopls", "check", str(file_path)], str(workspace)
        )
        diagnostics = parse_check_output(output)
        return make_payload(
            "diagnostics",
            {"file": str(file_path), "workspace": str(workspace)},
            {"diagnostics": diagnostics},
            str(workspace),
            result_count=len(diagnostics),
        )

    source = load_source(file_path)
    resolved = resolve_target(
        file_path, source, workspace, args.scope, getattr(args, "find", None), runner
    )
    target = {
        "file": str(file_path),
        "workspace": str(workspace),
        "scope": args.scope,
        "find": getattr(args, "find", None),
        "resolved": resolved,
    }

    if args.command == "locate":
        return make_payload(
            "locate",
            target,
            {"resolution": resolved},
            str(workspace),
        )

    cursor = resolved["cursor"]
    position = make_position(file_path, cursor["line"], cursor["column"])

    if args.command == "definition":
        output = invoke_runner(
            runner, ["gopls", "definition", "-json", position], str(workspace)
        )
        return make_payload(
            "definition",
            target,
            parse_definition(output),
            str(workspace),
            result_count=1,
        )

    if args.command == "references":
        output = invoke_runner(
            runner, ["gopls", "references", position], str(workspace)
        )
        references = parse_location_lines(output)
        return make_payload(
            "references",
            target,
            {"references": references},
            str(workspace),
            result_count=len(references),
        )

    if args.command == "implementation":
        output = invoke_runner(
            runner, ["gopls", "implementation", position], str(workspace)
        )
        implementations = parse_location_lines(output)
        return make_payload(
            "implementation",
            target,
            {"implementations": implementations},
            str(workspace),
            result_count=len(implementations),
        )

    if args.command == "call_hierarchy":
        output = invoke_runner(
            runner, ["gopls", "call_hierarchy", position], str(workspace)
        )
        result = parse_call_hierarchy(output)
        return make_payload(
            "call_hierarchy",
            target,
            result,
            str(workspace),
            caller_count=len(result["callers"]),
            callee_count=len(result["callees"]),
        )

    if args.command == "signature":
        output = invoke_runner(runner, ["gopls", "signature", position], str(workspace))
        signature = output.strip()
        return make_payload(
            "signature",
            target,
            {"signature": signature},
            str(workspace),
            result_count=1 if signature else 0,
        )

    if args.command == "highlight":
        output = invoke_runner(runner, ["gopls", "highlight", position], str(workspace))
        highlights = parse_location_lines(output)
        return make_payload(
            "highlight",
            target,
            {"highlights": highlights},
            str(workspace),
            result_count=len(highlights),
        )

    if args.command == "rename_validate":
        output = invoke_runner(
            runner, ["gopls", "prepare_rename", position], str(workspace)
        )
        return make_payload(
            "rename_validate",
            target,
            {"location": parse_single_location(output)},
            str(workspace),
            result_count=1,
        )

    if args.command == "rename_preview":
        output = invoke_runner(
            runner,
            ["gopls", "rename", "-d", position, args.name],
            str(workspace),
        )
        changed_files = extract_diff_files(output)
        return make_payload(
            "rename_preview",
            {**target, "new_name": args.name},
            {
                "mode": "diff",
                "changed_files": changed_files,
                "diff": output,
            },
            str(workspace),
            result_count=len(changed_files),
        )

    raise UserError("unsupported_command", f"Unsupported command: {args.command}")


def make_payload(command, target, result, workspace, **meta):
    payload = {
        "command": command,
        "target": target,
        "result": result,
        "meta": {
            "workspace": workspace,
        },
    }
    payload["meta"].update(meta)
    return payload


def resolve_target(file_path, source, workspace, scope, find_pattern, runner):
    scope_result = resolve_scope(file_path, source, workspace, scope, runner)
    resolved = {
        "scope_kind": scope_result["scope_kind"],
        "scope_range": scope_result["scope_range"],
        "range": scope_result["scope_range"],
        "cursor": scope_result["scope_range"]["start"],
    }
    if "symbol" in scope_result:
        resolved["symbol"] = scope_result["symbol"]

    if find_pattern is None:
        return resolved

    find_result = resolve_find(source, scope_result["scope_range"], find_pattern)
    resolved["range"] = find_result["range"]
    resolved["cursor"] = find_result["cursor"]
    return resolved


def resolve_scope(file_path, source, workspace, scope, runner):
    if LINE_SCOPE_RE.match(scope):
        line = int(scope)
        return {
            "scope_kind": "line",
            "scope_range": make_line_scope_range(source, line),
        }

    line_range_match = LINE_RANGE_SCOPE_RE.match(scope)
    if line_range_match:
        start = int(line_range_match.group("start"))
        end = int(line_range_match.group("end"))
        return {
            "scope_kind": "line_range",
            "scope_range": make_line_range_scope(source, start, end),
        }

    output = invoke_runner(runner, ["gopls", "symbols", str(file_path)], str(workspace))
    symbols = parse_outline_symbols(output, file_path)
    normalized_scope = normalize_symbol_path(scope)
    matches = [
        symbol
        for symbol in symbols
        if normalize_symbol_path(symbol["path"]) == normalized_scope
    ]
    if not matches:
        raise UserError("scope_not_found", f"Could not resolve scope: {scope}")
    if len(matches) > 1:
        raise UserError(
            "ambiguous_scope", f"Scope resolved to multiple symbols: {scope}"
        )
    match = matches[0]
    return {
        "scope_kind": "symbol",
        "scope_range": clone_range(match["range"]),
        "symbol": {
            "name": match["name"],
            "path": match["path"],
            "kind": match["kind"],
            "match": scope,
        },
    }


def resolve_find(source, scope_range, find_pattern):
    needle, marker_index = compact_pattern(find_pattern)
    if not needle:
        raise UserError(
            "find_not_found",
            "Find pattern does not contain any searchable non-whitespace characters.",
        )

    fragment, start_offset = extract_range_text(source, scope_range)
    haystack, mapping = compact_text(fragment, start_offset)
    matches = find_occurrences(haystack, needle)
    if not matches:
        raise UserError(
            "find_not_found",
            f"Could not find pattern inside the selected scope: {find_pattern}",
        )
    if len(matches) > 1:
        raise UserError(
            "ambiguous_find",
            f"Find pattern matched multiple matches inside the selected scope: {find_pattern}",
        )

    match_start = matches[0]
    match_start_offset = mapping[match_start]
    match_end_offset = mapping[match_start + len(needle) - 1] + 1
    result_range = {
        "start": offset_to_position(source, match_start_offset),
        "end": offset_to_position(source, match_end_offset),
    }

    if marker_index is None:
        cursor = result_range["start"]
    elif marker_index == len(needle):
        cursor = offset_to_position(source, match_end_offset)
    else:
        cursor = offset_to_position(source, mapping[match_start + marker_index])

    return {
        "range": result_range,
        "cursor": cursor,
    }


def compact_pattern(raw_pattern):
    if raw_pattern.count(CURSOR_MARKER) > 1:
        raise UserError(
            "invalid_cursor_marker",
            "Find pattern may contain at most one <|> cursor marker.",
        )

    compact_chars = []
    cursor_index = None
    index = 0
    while index < len(raw_pattern):
        if raw_pattern.startswith(CURSOR_MARKER, index):
            cursor_index = len(compact_chars)
            index += len(CURSOR_MARKER)
            continue

        character = raw_pattern[index]
        if not character.isspace():
            compact_chars.append(character)
        index += 1

    return "".join(compact_chars), cursor_index


def compact_text(text, start_offset):
    compact_chars = []
    offsets = []
    for index, character in enumerate(text):
        if character.isspace():
            continue
        compact_chars.append(character)
        offsets.append(start_offset + index)
    return "".join(compact_chars), offsets


def find_occurrences(haystack, needle):
    positions = []
    start = 0
    while True:
        match_index = haystack.find(needle, start)
        if match_index == -1:
            return positions
        positions.append(match_index)
        start = match_index + 1


def make_line_scope_range(source, line):
    if line < 1 or line > len(source["lines"]):
        raise UserError("invalid_scope", f"Line scope is outside the file: {line}")
    return {
        "start": {"line": line, "column": 1},
        "end": {"line": line, "column": len(source["lines"][line - 1]) + 1},
    }


def make_line_range_scope(source, start, end):
    if start <= 0 or end <= 0:
        raise UserError(
            "invalid_scope", "Line-range scope must use positive 1-based lines."
        )
    if start > end:
        raise UserError(
            "invalid_scope", "Line-range scope start must be less than or equal to end."
        )
    if end > len(source["lines"]):
        raise UserError(
            "invalid_scope", f"Line-range scope is outside the file: {start},{end}"
        )
    return {
        "start": {"line": start, "column": 1},
        "end": {"line": end, "column": len(source["lines"][end - 1]) + 1},
    }


def load_source(file_path):
    text = file_path.read_text(encoding="utf-8")
    raw_lines = text.splitlines(keepends=True) or [""]
    lines = []
    line_starts = []
    offset = 0
    for raw_line in raw_lines:
        line_starts.append(offset)
        line = raw_line.rstrip("\r\n")
        lines.append(line)
        offset += len(raw_line)
    return {
        "text": text,
        "lines": lines,
        "line_starts": line_starts,
    }


def extract_range_text(source, range_value):
    start_offset = position_to_offset(source, range_value["start"], "invalid_scope")
    end_offset = position_to_offset(source, range_value["end"], "invalid_scope")
    return source["text"][start_offset:end_offset], start_offset


def position_to_offset(source, position, error_code):
    line = position["line"]
    column = position["column"]
    if line < 1 or line > len(source["lines"]):
        raise UserError(error_code, f"Line is outside the file: {line}")
    max_column = len(source["lines"][line - 1]) + 1
    if column < 1 or column > max_column:
        raise UserError(error_code, f"Column is outside the file line: {line}:{column}")
    return source["line_starts"][line - 1] + column - 1


def offset_to_position(source, offset):
    for index, line_start in enumerate(source["line_starts"], start=1):
        line_length = len(source["lines"][index - 1])
        line_end = line_start + line_length
        if offset <= line_end:
            return {"line": index, "column": offset - line_start + 1}

    last_line = len(source["lines"])
    return {
        "line": last_line,
        "column": len(source["lines"][last_line - 1]) + 1,
    }


def clone_range(range_value):
    return {
        "start": dict(range_value["start"]),
        "end": dict(range_value["end"]),
    }


def normalize_workspace(raw_workspace):
    workspace = pathlib.Path(raw_workspace).expanduser().resolve()
    if not workspace.exists():
        raise UserError("workspace_not_found", f"Workspace does not exist: {workspace}")
    if not workspace.is_dir():
        raise UserError(
            "invalid_workspace", f"Workspace must be a directory: {workspace}"
        )
    return workspace


def normalize_file(raw_file):
    file_path = pathlib.Path(raw_file).expanduser().resolve()
    if not file_path.exists():
        raise UserError("file_not_found", f"Go file does not exist: {file_path}")
    if not file_path.is_file():
        raise UserError("invalid_file", f"Expected a file path, got: {file_path}")
    return file_path


def infer_workspace(file_path):
    current = file_path.parent
    for candidate in [current, *current.parents]:
        if any((candidate / marker).exists() for marker in WORKSPACE_MARKERS):
            return candidate
    return current


def make_position(file_path, line, column):
    if line <= 0 or column <= 0:
        raise UserError(
            "invalid_position", "Line and column must be positive 1-based integers."
        )
    return f"{file_path}:{line}:{column}"


def run_gopls(command, cwd):
    try:
        completed = subprocess.run(
            command,
            cwd=cwd,
            capture_output=True,
            text=True,
            check=False,
        )
    except FileNotFoundError as exc:
        raise UserError(
            "gopls_not_available",
            "gopls is not installed or not available in PATH.",
        ) from exc

    if completed.returncode != 0:
        stderr = (completed.stderr or completed.stdout or "").strip()
        raise classify_gopls_error(stderr, completed.returncode)

    return completed.stdout


def invoke_runner(runner, command, cwd):
    try:
        return runner(command, cwd)
    except FileNotFoundError as exc:
        raise UserError(
            "gopls_not_available",
            "gopls is not installed or not available in PATH.",
        ) from exc


def classify_gopls_error(stderr, returncode):
    message = stderr or "gopls exited with a non-zero status."
    lowered = message.lower()
    workspace_markers = (
        "go.mod file not found",
        "go.work file not found",
        "initial workspace load failed",
        "packages.load",
        "creating work dir",
    )
    if any(marker in lowered for marker in workspace_markers):
        return UserError(
            "workspace_load_failed",
            f"gopls could not load the Go workspace. Ensure the target is inside a module or go.work root. Original error: {message}",
            {"returncode": returncode},
        )
    if (
        "column is beyond end of line" in lowered
        or "line is beyond end of file" in lowered
    ):
        return UserError("invalid_position", message, {"returncode": returncode})
    return UserError(
        "gopls_command_failed",
        f"gopls command failed: {message}",
        {"returncode": returncode},
    )


def parse_definition(output):
    data = json.loads(output)
    return {
        "location": parse_uri_span(data.get("span", {})),
        "description": data.get("description", ""),
        "raw": data,
    }


def parse_workspace_symbols(output):
    symbols = []
    for line in non_empty_lines(output):
        match = SYMBOL_RE.match(line)
        if not match:
            symbols.append({"raw": line})
            continue
        symbols.append(
            {
                "location": parse_location(match.group("location")),
                "name": match.group("name"),
                "kind": match.group("kind"),
            }
        )
    return symbols


def parse_outline_symbols(output, file_path):
    symbols = []
    path_stack = []
    for raw_line in output.splitlines():
        line = raw_line.rstrip()
        if not line:
            continue
        match = OUTLINE_RE.match(line)
        if not match:
            symbols.append({"raw": line})
            continue

        depth = len(match.group("indent"))
        name = match.group("name")
        kind = match.group("kind")
        symbol_range = parse_line_range(match.group("range"))

        while len(path_stack) > depth:
            path_stack.pop()

        path = name if depth == 0 else f"{path_stack[depth - 1]}.{name}"
        if len(path_stack) == depth:
            path_stack.append(path)
        else:
            path_stack[depth] = path

        symbols.append(
            {
                "name": name,
                "kind": kind,
                "path": path,
                "range": {
                    "start": symbol_range["start"],
                    "end": symbol_range["end"],
                },
                "depth": depth,
                "location": {
                    "path": str(file_path),
                    "start": symbol_range["start"],
                    "end": symbol_range["end"],
                },
            }
        )
    return symbols


def normalize_symbol_path(path):
    segments = path.split(".")
    if len(segments) < 2:
        return path
    receiver = segments[0]
    if receiver.startswith("(") and receiver.endswith(")"):
        receiver = receiver[1:-1]
    segments[0] = receiver
    return ".".join(segments)


def parse_location_lines(output):
    return [parse_location(line) for line in non_empty_lines(output)]


def parse_single_location(output):
    lines = non_empty_lines(output)
    if len(lines) != 1:
        raise UserError(
            "unexpected_output", "Expected exactly one location in gopls output."
        )
    return parse_location(lines[0])


def parse_check_output(output):
    diagnostics = []
    for line in non_empty_lines(output):
        match = CHECK_RE.match(line)
        if not match:
            diagnostics.append({"raw": line})
            continue
        diagnostics.append(
            {
                "location": parse_location(match.group("location")),
                "message": match.group("message"),
            }
        )
    return diagnostics


def parse_call_hierarchy(output):
    result = {"identifier": None, "callers": [], "callees": [], "raw_lines": []}
    for line in non_empty_lines(output):
        identifier = CALL_IDENTIFIER_RE.match(line)
        if identifier:
            result["identifier"] = {
                "kind": identifier.group("symbol_kind"),
                "name": identifier.group("symbol_name"),
                "location": parse_location(identifier.group("symbol_location")),
            }
            continue

        edge = CALL_EDGE_RE.match(line)
        if edge:
            parsed = {
                "index": int(edge.group("index")),
                "selection": {
                    "path": edge.group("selection_path"),
                    "range": parse_line_range(edge.group("selection")),
                },
                "symbol": {
                    "kind": edge.group("symbol_kind"),
                    "name": edge.group("symbol_name"),
                    "location": parse_location(edge.group("symbol_location")),
                },
            }
            result[f"{edge.group('section')}s"].append(parsed)
            continue

        result["raw_lines"].append(line)
    return result


def parse_uri_span(span):
    uri = span.get("uri", "")
    parsed = urllib.parse.urlparse(uri)
    path = urllib.parse.unquote(parsed.path) if parsed.scheme == "file" else uri
    start = span.get("start", {})
    end = span.get("end", {})
    return {
        "path": path,
        "start": {
            "line": start.get("line"),
            "column": start.get("column"),
            "offset": start.get("offset"),
        },
        "end": {
            "line": end.get("line"),
            "column": end.get("column"),
            "offset": end.get("offset"),
        },
    }


def parse_location(raw):
    match = LOCATION_RE.match(raw)
    if not match:
        raise UserError(
            "unexpected_output", f"Could not parse location from gopls output: {raw}"
        )

    line = int(match.group("line"))
    column = int(match.group("column"))
    end_line = int(match.group("end_line") or line)
    end_column = int(match.group("end_column") or column)
    return {
        "path": match.group("path"),
        "start": {"line": line, "column": column},
        "end": {"line": end_line, "column": end_column},
    }


def parse_line_range(raw):
    location = parse_location(f"placeholder:{raw}")
    return {
        "start": location["start"],
        "end": location["end"],
    }


def extract_diff_files(output):
    changed_files = []
    for line in output.splitlines():
        if not line.startswith("+++ "):
            continue
        path = line[4:].strip()
        if path == "/dev/null":
            continue
        changed_files.append(path)
    return changed_files


def non_empty_lines(output):
    return [line.strip() for line in output.splitlines() if line.strip()]


def emit_error(exc):
    payload = {
        "error": {
            "code": exc.code,
            "message": exc.message,
        }
    }
    if exc.details:
        payload["error"]["details"] = exc.details
    json.dump(payload, sys.stderr, indent=2)
    sys.stderr.write("\n")


if __name__ == "__main__":
    sys.exit(main())
