from pathlib import Path
import shutil, subprocess, tempfile, re
root = Path.cwd()
artifact = Path(__file__).resolve().parent
with tempfile.TemporaryDirectory(prefix="snapid-review-", dir=root / ".temp") as tmp:
    dst = Path(tmp) / "probe"
    shutil.copytree(root / "internal/envsnapshot", dst)
    codec = dst / "codec_test.go"
    corpus = str(root / "internal/crossconformance/testdata/managed-home-snapshot-v1")
    codec.write_text(re.sub(r'const corpus = "[^"]+"', 'const corpus = "' + corpus + '"', codec.read_text()))
    shutil.copyfile(artifact / "TASK-261016-1v3gbc_review-extra-rev2_test.go", dst / "review_extra_test.go")
    result = subprocess.run(["go", "test", "-count=1", "-v", "-run", "^TestReviewAdditionalCaselessCollisions$", str(dst)])
    raise SystemExit(result.returncode)
