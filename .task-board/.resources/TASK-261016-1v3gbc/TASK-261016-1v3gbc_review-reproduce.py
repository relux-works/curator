from pathlib import Path
import shutil, subprocess, sys
root = Path.cwd()
source = root / "internal/envsnapshot"
target = root / ".temp/snapid-review-replay"
shutil.copytree(source, target, dirs_exist_ok=True)
p = target / "codec_test.go"
p.write_text(p.read_text().replace("../crossconformance/testdata/managed-home-snapshot-v1", str(root / "internal/crossconformance/testdata/managed-home-snapshot-v1")))
shutil.copyfile(Path(__file__).with_name("TASK-261016-1v3gbc_reviewer_test.go"), target / "reviewer_test.go")
raise SystemExit(subprocess.call(["go", "test", "-count=1", "-v", "-run", sys.argv[1] if len(sys.argv)>1 else "^TestReviewer", "./.temp/snapid-review-replay"]))
