#!/usr/bin/env python3
"""Positive controls plus exact refusal witnesses for every aggregate dependency."""
import importlib.util
from pathlib import Path
import unittest
import sys

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location("gate", Path(__file__).with_name("macos-policy-gate.py"))
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class PolicyTests(unittest.TestCase):
    # Independent event expectations include default PRs, label removal and train/main.
    def test_selection(self):
        for name, ref, labels, expected in (
            ("pull_request", "refs/pull/1/merge", [], False),
            ("pull_request", "refs/pull/1/merge", ["other"], False),
            ("pull_request", "refs/pull/1/merge", ["landing-candidate"], True),
            ("pull_request", "refs/pull/1/merge", ["LANDING-CANDIDATE"], True),
            ("push", "refs/heads/main", [], True),
            ("push", "refs/heads/gate/control", [], False),
            ("merge_group", "refs/heads/gh-readonly-queue/main/control", [], True),
            ("workflow_dispatch", "refs/heads/control", [], True),
        ):
            event = {"pull_request": {"labels": [{"name": label} for label in labels]}}
            self.assertEqual(gate.full_macos(name, ref, event), expected)

    # Every selected dependency must succeed: skipped/cancelled/missing are not green.
    # Each negative has a nearby all-success control and checks the named refusal.
    def test_dependency_verdicts(self):
        for name, ref, event in (
            ("pull_request", "refs/pull/1/merge", {"pull_request": {"labels": []}}),
            ("pull_request", "refs/pull/1/merge", {"pull_request": {"labels": [{"name": "landing-candidate"}]}}),
            ("push", "refs/heads/main", {}),
            ("merge_group", "refs/heads/gh-readonly-queue/main/control", {}),
            ("workflow_dispatch", "refs/heads/control", {}),
            ("workflow_dispatch", "refs/heads/control", {"inputs": {"candidate_ref": "a" * 40}}),
            ("workflow_dispatch", "refs/heads/control", {"inputs": {"candidate_root": "/fixture"}}),
        ):
            required = list(gate.MATRIX_JOBS)
            if name == "pull_request":
                required.append("macos-smoke")
            if event.get("inputs"):
                required.append("candidate-conformance")
            good = {job: {"result": "success"} for job in required}
            self.assertIn("passed", gate.check(name, ref, event, good))
            for job in required:
                for status in ("skipped", "cancelled", "failure", "missing"):
                    with self.subTest(event=name, job=job, status=status):
                        bad = dict(good)
                        if status == "missing":
                            del bad[job]
                        else:
                            bad[job] = {"result": status}
                        with self.assertRaisesRegex(ValueError, f"{job}: expected success, got {status}"):
                            gate.check(name, ref, event, bad)

    # Optional candidate suites and non-PR smoke can skip without attesting them.
    def test_optional_jobs(self):
        needs = {job: {"result": "success"} for job in gate.MATRIX_JOBS}
        needs.update({"macos-smoke": {"result": "skipped"}, "candidate-conformance": {"result": "skipped"}})
        self.assertIn("Full macOS matrix passed", gate.check("push", "refs/heads/main", {}, needs))
        needs["macos-smoke"] = {"result": "success"}
        self.assertIn("intentionally not selected", gate.check("pull_request", "refs/pull/1/merge", {"pull_request": {"labels": []}}, needs))


if __name__ == "__main__":
    unittest.main()
