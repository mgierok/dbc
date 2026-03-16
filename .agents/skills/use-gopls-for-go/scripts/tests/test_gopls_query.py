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


class GoplsQueryTests(unittest.TestCase):
    def test_definition_prints_normalized_json(self):
        output = json.dumps(
            {
                "span": {
                    "uri": "file:///tmp/sample.go",
                    "start": {"line": 3, "column": 5, "offset": 21},
                    "end": {"line": 3, "column": 12, "offset": 28},
                },
                "description": "type Example struct{}",
            }
        )

        def fake_runner(command, cwd):
            self.assertEqual(
                command,
                ["gopls", "definition", "-json", f"{FIXTURE_FILE}:3:5"],
            )
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return output

        stdout = io.StringIO()
        stderr = io.StringIO()
        with contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
            exit_code = MODULE.main(
                ["definition", "--file", str(FIXTURE_FILE), "--line", "3", "--column", "5"],
                runner=fake_runner,
            )

        self.assertEqual(exit_code, 0)
        self.assertEqual(stderr.getvalue(), "")
        payload = json.loads(stdout.getvalue())
        self.assertEqual(payload["command"], "definition")
        self.assertEqual(payload["target"]["workspace"], str(FIXTURE_WORKSPACE))
        self.assertEqual(payload["result"]["location"]["path"], "/tmp/sample.go")
        self.assertEqual(payload["result"]["description"], "type Example struct{}")

    def test_rename_diff_uses_diff_flag_without_write(self):
        def fake_runner(command, cwd):
            self.assertEqual(
                command,
                ["gopls", "rename", "-d", f"{FIXTURE_FILE}:8:2", "RenamedSymbol"],
            )
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return (
                f"--- {FIXTURE_FILE}.orig\n"
                f"+++ {FIXTURE_FILE}\n"
                "@@ -1,2 +1,2 @@\n"
                "-type OldName struct{}\n"
                "+type RenamedSymbol struct{}\n"
            )

        stdout = io.StringIO()
        with contextlib.redirect_stdout(stdout):
            exit_code = MODULE.main(
                [
                    "rename_diff",
                    "--file",
                    str(FIXTURE_FILE),
                    "--line",
                    "8",
                    "--column",
                    "2",
                    "--name",
                    "RenamedSymbol",
                ],
                runner=fake_runner,
            )

        self.assertEqual(exit_code, 0)
        payload = json.loads(stdout.getvalue())
        self.assertEqual(payload["command"], "rename_diff")
        self.assertEqual(payload["result"]["mode"], "diff")
        self.assertEqual(payload["result"]["changed_files"], [str(FIXTURE_FILE)])

    def test_missing_gopls_returns_actionable_error(self):
        def missing_runner(command, cwd):
            raise FileNotFoundError("gopls")

        stdout = io.StringIO()
        stderr = io.StringIO()
        with contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
            exit_code = MODULE.main(
                ["check", "--file", str(FIXTURE_FILE)],
                runner=missing_runner,
            )

        self.assertEqual(exit_code, 1)
        self.assertEqual(stdout.getvalue(), "")
        payload = json.loads(stderr.getvalue())
        self.assertEqual(payload["error"]["code"], "gopls_not_available")
        self.assertIn("PATH", payload["error"]["message"])

    def test_invalid_column_returns_validation_error(self):
        stderr = io.StringIO()
        with contextlib.redirect_stderr(stderr):
            exit_code = MODULE.main(
                ["references", "--file", str(FIXTURE_FILE), "--line", "3", "--column", "0"]
            )

        self.assertEqual(exit_code, 1)
        payload = json.loads(stderr.getvalue())
        self.assertEqual(payload["error"]["code"], "invalid_position")
        self.assertIn("positive", payload["error"]["message"])

    def test_workspace_symbol_parses_symbol_lines(self):
        def fake_runner(command, cwd):
            self.assertEqual(command, ["gopls", "workspace_symbol", "MySymbol"])
            self.assertEqual(cwd, str(FIXTURE_WORKSPACE))
            return (
                f"{FIXTURE_WORKSPACE / 'main.go'}:10:6-14 MySymbol Function\n"
                f"{FIXTURE_WORKSPACE / 'lib.go'}:4:6-12 MyType Struct\n"
            )

        stdout = io.StringIO()
        with contextlib.redirect_stdout(stdout):
            exit_code = MODULE.main(
                ["workspace_symbol", "--workspace", str(FIXTURE_WORKSPACE), "--query", "MySymbol"],
                runner=fake_runner,
            )

        self.assertEqual(exit_code, 0)
        payload = json.loads(stdout.getvalue())
        self.assertEqual(payload["target"]["workspace"], str(FIXTURE_WORKSPACE))
        self.assertEqual(len(payload["result"]["symbols"]), 2)
        self.assertEqual(payload["result"]["symbols"][0]["name"], "MySymbol")
        self.assertEqual(payload["result"]["symbols"][1]["kind"], "Struct")

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
