# Reviewer reproduction

Candidate tree: 43b7ed0e9efd36ffcc321f6604d72aea221e2af7.
Download TASK-261016-hzbo83_review-attacks.go into .temp/review-hzbo83/attacks.go.
The attacks add tests only via a Go overlay; all production blobs stay pinned.
Run from the candidate checkout:

```sh
python3 - <<'PY'
from pathlib import Path
import json
root = Path.cwd()
base = root / '.temp/review-hzbo83'
(base / 'attacks-overlay.json').write_text(json.dumps({'Replace': {
    str(root / 'internal/envprofile/review_hzbo83_test.go'): str(base / 'attacks.go')
}}))
PY
go test -overlay .temp/review-hzbo83/attacks-overlay.json ./internal/envprofile -run '^TestReview(TransientPolicyRead|StatusBackendSwitch|DetachedBackendSwitch)$' -count=1 -v
```

Expected result: exit 0. Each test names the refused condition, unchanged
protected snapshot, and positive control. These are review evidence, not
committed regression tests or a claim of exhaustive coverage.

Narrowing controls are the attached TASK-261016-hzbo83_controls.py. The reviewer
ran rows M1-M6 and M7 separately while walking the corresponding surface rows;
only their scratch-output directory changed. All seven child commands exited
1 with the expected named behavioral test failure. See review-local-evidence.

The hosted review gate used a scratch copy of scripts/remote-gate.sh. The only
behavioral changes to gate orchestration were: replace its unsigned synthetic
author commit-tree with configured-human signed `git commit-tree -S`, verify
the commit signature, and assert its tree equals the candidate before pushing.
Workflow, jobs, test commands, retry and cleanup behavior stayed unchanged.
The Story branch, index and HEAD were not committed or moved. The synthetic
signed gate commit is 04e918e4e938e274bbcf9835a97a0b17b18122de.
