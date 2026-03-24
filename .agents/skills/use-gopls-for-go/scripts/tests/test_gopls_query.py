import contextlib
import importlib.util
import io
import json
import pathlib
import unittest


SCRIPT_PATH = pathlib.Path(__file__).resolve().parents[1] / "gopls_query.py"
FIXTURE_WORKSPACE = pathlib.Path(__file__).resolve().parent / "fixtures" / "workspace"
FIXTURE_FILE = FIXTURE_WORKSPACE / "main.go"
SPEC = importlib.util.spec_from_file_location("gopls_query", SCRIPT_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


def run_main(argv, runner=None):
    stdout = io.StringIO()
    stderr = io.StringIO()
    with contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
        exit_code = MODULE.main(argv, runner=runner)
    return exit_code, stdout.getvalue(), stderr.getvalue()


class GoplsQueryTests(unittest.TestCase):
    def test_search_prints_normalized_json(self):
        def fake_runner(command, cwd):
            self.assertEqual(command, ["gopls", "workspace_symbol", "OldName"])
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return (
                f"{FIXTURE_FILE}:7:6-13 OldName Struct\n"
                f"{FIXTURE_FILE}:21:6-16 UseOldName Function\n"
            )

        exit_code, stdout, stderr = run_main(
            ["search", "--workspace", str(FIXTURE_WORKSPACE), "--query", "OldName"],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "search")
        self.assertEqual(payload["target"]["workspace"], str(FIXTURE_WORKSPACE))
        self.assertEqual(payload["target"]["query"], "OldName")
        self.assertEqual(payload["meta"]["workspace"], str(FIXTURE_WORKSPACE))
        self.assertEqual(payload["meta"]["result_count"], 2)
        self.assertEqual(payload["result"]["symbols"][0]["name"], "OldName")
        self.assertEqual(payload["result"]["symbols"][1]["kind"], "Function")

    def test_outline_parses_symbol_depth_and_range(self):
        def fake_runner(command, cwd):
            self.assertEqual(command, ["gopls", "symbols", str(FIXTURE_FILE)])
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return (
                "Worker Interface 3:6-3:12\n"
                "\tWork Method 4:2-4:6\n"
                "OldName Struct 7:6-7:13\n"
                "(OldName).Work Method 9:16-9:20\n"
            )

        exit_code, stdout, stderr = run_main(
            ["outline", "--file", str(FIXTURE_FILE)],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "outline")
        self.assertEqual(payload["meta"]["result_count"], 4)
        self.assertEqual(payload["result"]["symbols"][0]["name"], "Worker")
        self.assertEqual(payload["result"]["symbols"][1]["depth"], 1)
        self.assertEqual(payload["result"]["symbols"][1]["path"], "Worker.Work")
        self.assertEqual(payload["result"]["symbols"][2]["kind"], "Struct")
        self.assertEqual(payload["result"]["symbols"][3]["path"], "(OldName).Work")

    def test_locate_resolves_single_line_scope_without_runner(self):
        exit_code, stdout, stderr = run_main(
            ["locate", "--file", str(FIXTURE_FILE), "--scope", "21"]
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        resolved = payload["target"]["resolved"]
        self.assertEqual(payload["command"], "locate")
        self.assertEqual(payload["meta"]["workspace"], str(FIXTURE_WORKSPACE))
        self.assertEqual(resolved["scope_kind"], "line")
        self.assertEqual(resolved["cursor"], {"line": 21, "column": 1})
        self.assertEqual(resolved["scope_range"]["start"], {"line": 21, "column": 1})
        self.assertEqual(resolved["scope_range"]["end"], {"line": 21, "column": 40})
        self.assertEqual(resolved["range"], resolved["scope_range"])
        self.assertEqual(payload["result"]["resolution"], resolved)

    def test_locate_resolves_line_range_find_with_cursor_marker(self):
        exit_code, stdout, stderr = run_main(
            [
                "locate",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "21,22",
                "--find",
                "Use<|>OldName",
            ]
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        resolved = payload["target"]["resolved"]
        self.assertEqual(resolved["scope_kind"], "line_range")
        self.assertEqual(resolved["cursor"], {"line": 21, "column": 9})
        self.assertEqual(resolved["range"]["start"], {"line": 21, "column": 6})
        self.assertEqual(resolved["range"]["end"], {"line": 21, "column": 16})
        self.assertEqual(resolved["scope_range"]["start"], {"line": 21, "column": 1})
        self.assertEqual(resolved["scope_range"]["end"], {"line": 22, "column": 25})

    def test_locate_find_is_whitespace_insensitive(self):
        exit_code, stdout, stderr = run_main(
            [
                "locate",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "17,18",
                "--find",
                "return worker. Work()",
            ]
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        resolved = payload["target"]["resolved"]
        self.assertEqual(resolved["cursor"], {"line": 18, "column": 2})
        self.assertEqual(resolved["range"]["start"], {"line": 18, "column": 2})
        self.assertEqual(resolved["range"]["end"], {"line": 18, "column": 22})

    def test_locate_resolves_symbol_scope_with_receiver_alias(self):
        def fake_runner(command, cwd):
            self.assertEqual(command, ["gopls", "symbols", str(FIXTURE_FILE)])
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return (
                "Worker Interface 3:6-3:12\n"
                "\tWork Method 4:2-4:6\n"
                "OldName Struct 7:6-7:13\n"
                "(OldName).Work Method 9:16-9:20\n"
            )

        exit_code, stdout, stderr = run_main(
            ["locate", "--file", str(FIXTURE_FILE), "--scope", "OldName.Work"],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        resolved = payload["target"]["resolved"]
        self.assertEqual(resolved["scope_kind"], "symbol")
        self.assertEqual(resolved["cursor"], {"line": 9, "column": 16})
        self.assertEqual(resolved["scope_range"]["start"], {"line": 9, "column": 16})
        self.assertEqual(resolved["scope_range"]["end"], {"line": 9, "column": 20})
        self.assertEqual(resolved["symbol"]["name"], "(OldName).Work")
        self.assertEqual(resolved["symbol"]["path"], "(OldName).Work")
        self.assertEqual(resolved["symbol"]["match"], "OldName.Work")

    def test_definition_uses_resolved_cursor_from_symbol_scope_and_find(self):
        symbols_output = "OldName Struct 7:6-7:13\n(OldName).Work Method 9:16-9:20\n"
        definition_output = json.dumps(
            {
                "span": {
                    "uri": "file:///tmp/helper.go",
                    "start": {"line": 13, "column": 6, "offset": 111},
                    "end": {"line": 13, "column": 12, "offset": 117},
                },
                "description": "func helper() string",
            }
        )

        def fake_runner(command, cwd):
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            if command == ["gopls", "symbols", str(FIXTURE_FILE)]:
                return symbols_output
            self.assertEqual(
                command,
                ["gopls", "definition", "-json", f"{FIXTURE_FILE}:9:18"],
            )
            return definition_output

        exit_code, stdout, stderr = run_main(
            [
                "definition",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "OldName.Work",
                "--find",
                "Wo<|>rk",
            ],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "definition")
        self.assertEqual(
            payload["target"]["resolved"]["cursor"], {"line": 9, "column": 18}
        )
        self.assertEqual(payload["result"]["location"]["path"], "/tmp/helper.go")
        self.assertEqual(payload["result"]["description"], "func helper() string")

    def test_signature_returns_text_result(self):
        def fake_runner(command, cwd):
            self.assertEqual(
                command,
                ["gopls", "signature", f"{FIXTURE_FILE}:18:16"],
            )
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return "Work() string\n"

        exit_code, stdout, stderr = run_main(
            [
                "signature",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "18",
                "--find",
                "worker.<|>Work",
            ],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "signature")
        self.assertEqual(payload["result"]["signature"], "Work() string")

    def test_highlight_parses_location_list(self):
        def fake_runner(command, cwd):
            self.assertEqual(
                command,
                ["gopls", "highlight", f"{FIXTURE_FILE}:21:22"],
            )
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return f"{FIXTURE_FILE}:21:17-22\n{FIXTURE_FILE}:22:19-24\n"

        exit_code, stdout, stderr = run_main(
            [
                "highlight",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "21",
                "--find",
                "value<|>",
            ],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["meta"]["result_count"], 2)
        self.assertEqual(payload["result"]["highlights"][0]["start"]["column"], 17)
        self.assertEqual(payload["result"]["highlights"][1]["end"]["column"], 24)

    def test_diagnostics_parses_output(self):
        def fake_runner(command, cwd):
            self.assertEqual(command, ["gopls", "check", str(FIXTURE_FILE)])
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return f"{FIXTURE_FILE}:7:6-13: example diagnostic\n"

        exit_code, stdout, stderr = run_main(
            ["diagnostics", "--file", str(FIXTURE_FILE)],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "diagnostics")
        self.assertEqual(payload["meta"]["result_count"], 1)
        self.assertEqual(
            payload["result"]["diagnostics"][0]["message"], "example diagnostic"
        )

    def test_rename_validate_uses_prepare_rename(self):
        def fake_runner(command, cwd):
            self.assertEqual(
                command,
                ["gopls", "prepare_rename", f"{FIXTURE_FILE}:7:9"],
            )
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return f"{FIXTURE_FILE}:7:6-13\n"

        exit_code, stdout, stderr = run_main(
            [
                "rename_validate",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "7",
                "--find",
                "Old<|>Name",
            ],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "rename_validate")
        self.assertEqual(payload["result"]["location"]["start"]["column"], 6)

    def test_rename_preview_uses_diff_flag_without_write(self):
        def fake_runner(command, cwd):
            self.assertEqual(
                command,
                ["gopls", "rename", "-d", f"{FIXTURE_FILE}:7:9", "RenamedSymbol"],
            )
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return (
                f"--- {FIXTURE_FILE}.orig\n"
                f"+++ {FIXTURE_FILE}\n"
                "@@ -1,2 +1,2 @@\n"
                "-type OldName struct{}\n"
                "+type RenamedSymbol struct{}\n"
            )

        exit_code, stdout, stderr = run_main(
            [
                "rename_preview",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "7",
                "--find",
                "Old<|>Name",
                "--name",
                "RenamedSymbol",
            ],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr, "")
        payload = json.loads(stdout)
        self.assertEqual(payload["command"], "rename_preview")
        self.assertEqual(payload["target"]["new_name"], "RenamedSymbol")
        self.assertEqual(payload["result"]["mode"], "diff")
        self.assertEqual(payload["result"]["changed_files"], [str(FIXTURE_FILE)])

    def test_ambiguous_find_returns_actionable_error(self):
        exit_code, stdout, stderr = run_main(
            [
                "locate",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "1,23",
                "--find",
                "Work",
            ]
        )

        self.assertEqual(exit_code, 1)
        self.assertEqual(stdout, "")
        payload = json.loads(stderr)
        self.assertEqual(payload["error"]["code"], "ambiguous_find")
        self.assertIn("multiple matches", payload["error"]["message"])

    def test_scope_not_found_returns_actionable_error(self):
        def fake_runner(command, cwd):
            self.assertEqual(command, ["gopls", "symbols", str(FIXTURE_FILE)])
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return "OldName Struct 7:6-7:13\n"

        exit_code, stdout, stderr = run_main(
            ["locate", "--file", str(FIXTURE_FILE), "--scope", "Missing.Symbol"],
            runner=fake_runner,
        )

        self.assertEqual(exit_code, 1)
        self.assertEqual(stdout, "")
        payload = json.loads(stderr)
        self.assertEqual(payload["error"]["code"], "scope_not_found")
        self.assertIn("Missing.Symbol", payload["error"]["message"])

    def test_invalid_cursor_marker_returns_actionable_error(self):
        exit_code, stdout, stderr = run_main(
            [
                "locate",
                "--file",
                str(FIXTURE_FILE),
                "--scope",
                "7",
                "--find",
                "Old<|><|>Name",
            ]
        )

        self.assertEqual(exit_code, 1)
        self.assertEqual(stdout, "")
        payload = json.loads(stderr)
        self.assertEqual(payload["error"]["code"], "invalid_cursor_marker")

    def test_missing_gopls_returns_actionable_error(self):
        def missing_runner(command, cwd):
            raise FileNotFoundError("gopls")

        exit_code, stdout, stderr = run_main(
            ["diagnostics", "--file", str(FIXTURE_FILE)],
            runner=missing_runner,
        )

        self.assertEqual(exit_code, 1)
        self.assertEqual(stdout, "")
        payload = json.loads(stderr)
        self.assertEqual(payload["error"]["code"], "gopls_not_available")
        self.assertIn("PATH", payload["error"]["message"])

    def test_workspace_load_failure_takes_precedence_over_position_error(self):
        error = MODULE.classify_gopls_error(
            "initial workspace load failed: go/packages.Load: creating work dir: operation not permitted\n"
            "gopls: start: column is beyond end of line",
            2,
        )

        self.assertEqual(error.code, "workspace_load_failed")
        self.assertIn("could not load the Go workspace", error.message)


if __name__ == "__main__":
    unittest.main()
