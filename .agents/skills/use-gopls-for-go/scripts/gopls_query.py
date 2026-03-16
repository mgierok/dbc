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
WORKSPACE_MARKERS = ("go.work", "go.mod")


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

    workspace_symbol = subparsers.add_parser("workspace_symbol")
    workspace_symbol.add_argument("--workspace", required=True)
    workspace_symbol.add_argument("--query", required=True)

    for name in (
        "definition",
        "references",
        "implementation",
        "call_hierarchy",
        "prepare_rename",
    ):
        file_parser = subparsers.add_parser(name)
        add_file_position_args(file_parser)

    check = subparsers.add_parser("check")
    check.add_argument("--file", required=True)

    rename_diff = subparsers.add_parser("rename_diff")
    add_file_position_args(rename_diff)
    rename_diff.add_argument("--name", required=True)

    return parser


def add_file_position_args(parser):
    parser.add_argument("--file", required=True)
    parser.add_argument("--line", required=True, type=int)
    parser.add_argument("--column", required=True, type=int)


def main(argv=None, runner=None):
    args = build_parser().parse_args(argv)
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
    if args.command == "workspace_symbol":
        workspace = normalize_workspace(args.workspace)
        output = invoke_runner(runner, ["gopls", "workspace_symbol", args.query], str(workspace))
        return {
            "command": args.command,
            "target": {"workspace": str(workspace), "query": args.query},
            "result": {"symbols": parse_workspace_symbols(output)},
        }

    file_path = normalize_file(args.file)
    workspace = infer_workspace(file_path)
    target = {"file": str(file_path), "workspace": str(workspace)}

    if args.command == "check":
        output = invoke_runner(runner, ["gopls", "check", str(file_path)], str(workspace))
        return {
            "command": args.command,
            "target": target,
            "result": {"diagnostics": parse_check_output(output)},
        }

    position = make_position(file_path, args.line, args.column)
    target.update({"line": args.line, "column": args.column})

    if args.command == "definition":
        output = invoke_runner(runner, ["gopls", "definition", "-json", position], str(workspace))
        return {
            "command": args.command,
            "target": target,
            "result": parse_definition(output),
        }

    if args.command == "references":
        output = invoke_runner(runner, ["gopls", "references", position], str(workspace))
        return {
            "command": args.command,
            "target": target,
            "result": {"references": parse_location_lines(output)},
        }

    if args.command == "implementation":
        output = invoke_runner(runner, ["gopls", "implementation", position], str(workspace))
        return {
            "command": args.command,
            "target": target,
            "result": {"implementations": parse_location_lines(output)},
        }

    if args.command == "call_hierarchy":
        output = invoke_runner(runner, ["gopls", "call_hierarchy", position], str(workspace))
        return {
            "command": args.command,
            "target": target,
            "result": parse_call_hierarchy(output),
        }

    if args.command == "prepare_rename":
        output = invoke_runner(runner, ["gopls", "prepare_rename", position], str(workspace))
        return {
            "command": args.command,
            "target": target,
            "result": {"location": parse_single_location(output)},
        }

    if args.command == "rename_diff":
        output = invoke_runner(runner, ["gopls", "rename", "-d", position, args.name], str(workspace))
        return {
            "command": args.command,
            "target": {**target, "new_name": args.name},
            "result": {
                "mode": "diff",
                "changed_files": extract_diff_files(output),
                "diff": output,
            },
        }

    raise UserError("unsupported_command", f"Unsupported command: {args.command}")


def normalize_workspace(raw_workspace):
    workspace = pathlib.Path(raw_workspace).expanduser().resolve()
    if not workspace.exists():
        raise UserError("workspace_not_found", f"Workspace does not exist: {workspace}")
    if not workspace.is_dir():
        raise UserError("invalid_workspace", f"Workspace must be a directory: {workspace}")
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
        raise UserError("invalid_position", "Line and column must be positive 1-based integers.")
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
    if "column is beyond end of line" in lowered or "line is beyond end of file" in lowered:
        return UserError("invalid_position", message, {"returncode": returncode})
    return UserError(
        "gopls_command_failed",
        f"gopls command failed: {message}",
        {"returncode": returncode},
    )


def parse_definition(output):
    data = json.loads(output)
    location = parse_uri_span(data.get("span", {}))
    return {
        "location": location,
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


def parse_location_lines(output):
    return [parse_location(line) for line in non_empty_lines(output)]


def parse_single_location(output):
    lines = non_empty_lines(output)
    if len(lines) != 1:
        raise UserError("unexpected_output", "Expected exactly one location in gopls output.")
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
        raise UserError("unexpected_output", f"Could not parse location from gopls output: {raw}")

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
