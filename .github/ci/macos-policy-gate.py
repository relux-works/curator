#!/usr/bin/env python3
"""Fail closed on missing coverage; a smoke verdict never attests full macOS CI."""
import json
import os
import sys

MATRIX_JOBS = (
    "go-family-qualification", "unix-test-shards", "test",
    "envprofile-race", "race", "gate-selftest",
)


def full_macos(event_name, ref, event):
    return (
        event_name in ("workflow_dispatch", "merge_group")
        or (event_name == "push" and ref == "refs/heads/main")
        or (event_name == "pull_request" and any(
            label["name"].lower() == "landing-candidate"
            for label in event["pull_request"]["labels"]
        ))
    )


def check(event_name, ref, event, needs):
    full = full_macos(event_name, ref, event)
    # Mixed matrices still contain their unchanged Linux/Windows entries on PRs.
    # The workflow controls prove macOS entries are present exactly when full.
    required = list(MATRIX_JOBS)
    if event_name == "pull_request":
        required.append("macos-smoke")
    inputs = event.get("inputs") or {}
    if event_name == "workflow_dispatch" and (
        inputs.get("candidate_ref") or inputs.get("candidate_root")
    ):
        required.append("candidate-conformance")
    for name in required:
        result = needs.get(name, {}).get("result", "missing")
        if result != "success":
            raise ValueError(f"{name}: expected success, got {result}")
    if full:
        return "Full macOS matrix passed (plus smoke on pull requests)."
    if event_name == "pull_request":
        return "PR smoke passed; full macOS matrix was intentionally not selected. Apply landing-candidate before landing."
    return "Full macOS matrix not selected on this non-landing ref."


if __name__ == "__main__":
    try:
        with open(os.environ["GITHUB_EVENT_PATH"], encoding="utf-8") as stream:
            event = json.load(stream)
        verdict = check(os.environ["GITHUB_EVENT_NAME"], os.environ["GITHUB_REF"],
                        event, json.loads(os.environ["CI_NEEDS"]))
        print(verdict)
        summary = os.environ.get("GITHUB_STEP_SUMMARY")
        if summary:
            with open(summary, "a", encoding="utf-8") as stream:
                stream.write(verdict + "\n")
    except (KeyError, TypeError, ValueError, OSError) as error:
        sys.exit(f"macOS policy gate: {error}")
