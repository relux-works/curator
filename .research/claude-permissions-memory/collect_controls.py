#!/usr/bin/env python3
"""Positive, planted-violation and narrowing controls for publication filtering."""
import json
from pathlib import Path
import subprocess
import sys

root = Path(sys.argv[1])
if not root.is_absolute() or ".temp" not in root.parts or root.exists():
    raise SystemExit("fresh absolute task-local .temp control root required")
evidence = root / "input" / "TestSynthetic" / "evidence"
evidence.mkdir(parents=True)
(evidence / "identity.json").write_text(json.dumps({"version": "synthetic collector control"}))
source = Path(__file__).with_name("collect.py")
text = source.read_text()
lines = text.splitlines()
guards = [i for i, line in enumerate(lines) if line.startswith("if re.search(")]
if len(guards) != 1:
    raise SystemExit("publication guard plant was not applied exactly once")
lines[guards[0]] = "if False:  # bounded checker mutant, synthetic inputs only"
mutant = root / "collector-mutant.py"
mutant.write_text("\n".join(lines) + "\n")


def invoke(label, collector, locator):
    item = {"argv": [locator], "exit": 0, "requests": [{"messages": []}],
            "init": {"tools": []}, "result": {"permission_denials": []}}
    (evidence / "turn.json").write_text(json.dumps(item))
    output = root / (label + ".json")
    run = subprocess.run([sys.executable, str(collector), str(root / "input"), str(output)],
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    return run.returncode, run.stdout, output.exists()


# A sanitized locator must publish; each synthetic ordinary/flattened home
# locator must fail specifically before publishing. Disabling only that guard
# must expose each same violation, establishing the checker is discriminating.
def require(condition, message):
    if not condition:
        raise SystemExit(message)


code, body, exists = invoke("positive", source, "$RUN/home/.claude/managed.json")
require(code == 0 and exists and "records=1" in body, "sanitized positive did not publish")
print(json.dumps({"control": "sanitized-positive", "exit": code, "published": exists}))
plants = {"ordinary-macos": "/Users/fixture/private", "ordinary-linux": "/home/fixture/private",
          "flattened": "-Users-fixture-private"}
for label, locator in plants.items():
    code, body, exists = invoke(label, source, locator)
    require(code == 1 and not exists and "publication contains" in body, "wrong rejection for " + label)
    print(json.dumps({"control": label, "exit": code, "published": exists, "intended_rejection": True}))
    code, body, exists = invoke(label + "-mutant", mutant, locator)
    require(code == 0 and exists and "records=1" in body, "narrowing plant did not expose " + label)
    print(json.dumps({"control": label + "-mutant", "exit": code, "checker_violation_detected": True}))
