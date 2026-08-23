from __future__ import annotations

import json
import re
import subprocess
import sys
from pathlib import Path


if len(sys.argv) != 4:
    raise SystemExit(
        "usage: measure_changed_coverage.py OLD_SOURCE NEW_SOURCE COVERAGE_JSON"
    )

old_source = Path(sys.argv[1])
new_source = Path(sys.argv[2])
coverage_json = Path(sys.argv[3])
diff = subprocess.run(
    [
        "git",
        "diff",
        "--no-index",
        "--unified=0",
        "--",
        str(old_source),
        str(new_source),
    ],
    check=False,
    capture_output=True,
    text=True,
)
if diff.returncode != 1:
    raise SystemExit(
        f"expected a changed-file git diff exit 1, got {diff.returncode}"
    )

hunk = re.compile(r"^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@")
added_lines: set[int] = set()
new_line: int | None = None
for line in diff.stdout.splitlines():
    match = hunk.match(line)
    if match is not None:
        new_line = int(match.group(1))
        continue
    if new_line is None or line.startswith(("---", "+++")):
        continue
    if line.startswith("+"):
        added_lines.add(new_line)
        new_line += 1
    elif line.startswith("-"):
        continue
    elif line.startswith(" "):
        new_line += 1

report = json.loads(coverage_json.read_text(encoding="utf-8"))
files = report["files"]
try:
    file_report = files["src/csk/builds/go_v1.py"]
except KeyError as exc:
    raise SystemExit("coverage report has no go_v1.py entry") from exc

executed_lines = set(file_report["executed_lines"])
missing_lines = set(file_report["missing_lines"])
executable_lines = executed_lines | missing_lines
changed_executable = added_lines & executable_lines
covered_changed = changed_executable & executed_lines
missing_changed = changed_executable & missing_lines

executed_branches = {
    tuple(branch) for branch in file_report["executed_branches"]
}
missing_branches = {
    tuple(branch) for branch in file_report["missing_branches"]
}
all_branches = executed_branches | missing_branches
changed_branches = {
    branch for branch in all_branches if branch[0] in added_lines
}
covered_changed_branches = changed_branches & executed_branches
missing_changed_branches = changed_branches & missing_branches


def percentage(covered: int, total: int) -> float:
    return 100.0 if total == 0 else covered * 100.0 / total


line_percent = percentage(len(covered_changed), len(changed_executable))
branch_percent = percentage(
    len(covered_changed_branches),
    len(changed_branches),
)
combined_covered = len(covered_changed) + len(covered_changed_branches)
combined_total = len(changed_executable) + len(changed_branches)
combined_percent = percentage(combined_covered, combined_total)

print(f"old_source={old_source}")
print(f"new_source={new_source}")
print("diff_exit=1 (expected: the reviewed and cycle-5 sources differ)")
print(f"added_physical_lines={len(added_lines)}")
print(
    "changed_line_coverage="
    f"{len(covered_changed)}/{len(changed_executable)} "
    f"({line_percent:.2f}%)"
)
print(
    "changed_branch_coverage="
    f"{len(covered_changed_branches)}/{len(changed_branches)} "
    f"({branch_percent:.2f}%)"
)
print(
    "changed_combined_coverage="
    f"{combined_covered}/{combined_total} ({combined_percent:.2f}%)"
)
print(
    "missing_changed_lines="
    + ",".join(str(line) for line in sorted(missing_changed))
)
print(
    "missing_changed_branches="
    + ",".join(
        f"{source}->{target}"
        for source, target in sorted(missing_changed_branches)
    )
)
