import re
import subprocess
import unittest
from pathlib import Path
from urllib.parse import unquote, urlsplit

ROOT = Path.cwd()
REV1 = "891d805d"
OLD_BASE = "bab2433b"
NEW_BASE = "bd126a9a"
GUIDE = "docs/second-operator.md"

def blob(rev, path):
    return subprocess.check_output(["git", "show", f"{rev}:{path}"]).decode()

def expected_readme():
    old = blob(OLD_BASE, "README.md").splitlines(keepends=True)
    new = blob(NEW_BASE, "README.md").splitlines(keepends=True)
    before = [line for line in old if line.startswith("- **Operator credentials**:")]
    after = [line for line in new if line.startswith("- **Operator credentials**:")]
    assert len(before) == len(after) == 1
    accepted = blob(REV1, "README.md")
    assert accepted.count(before[0]) == 1
    return accepted.replace(before[0], after[0])

def headings(text):
    seen = {}
    result = set()
    for title in re.findall(r"^#{1,6}\s+(.+?)\s*#*\s*$", text, re.M):
        title = re.sub(r"<[^>]+>", "", title).lower()
        slug = re.sub(r"[^\w\- ]", "", title).replace(" ", "-")
        count = seen.get(slug, 0)
        seen[slug] = count + 1
        result.add(slug if count == 0 else f"{slug}-{count}")
    return result

def relative_links(path, text):
    text = re.sub(r"^```.*?^```[^\n]*$", "", text, flags=re.M | re.S)
    links = re.findall(r"\[[^\]]*\]\(([^)]+)\)", text)
    relative = []
    for target in links:
        target = target.strip().strip("<>")
        url = urlsplit(target)
        if url.scheme or url.netloc:
            continue
        resolved = ROOT / Path(path).parent / unquote(url.path) if url.path else ROOT / path
        if not resolved.exists():
            raise AssertionError(f"{path}: missing {target}")
        if url.fragment:
            if not resolved.is_file() or unquote(url.fragment) not in headings(resolved.read_text()):
                raise AssertionError(f"{path}: missing anchor {target}")
        relative.append(target)
    return relative

class RefreshRegression(unittest.TestCase):
    def test_only_carrier_paths_change(self):
        paths = subprocess.check_output(["git", "diff", "--name-only", NEW_BASE]).decode().splitlines()
        self.assertEqual(set(paths), {"README.md", GUIDE})
        state = subprocess.check_output(["git", "status", "--porcelain", "-z"]).decode().split("\0")
        self.assertEqual({row[3:] for row in state if row}, {"README.md", GUIDE})

    def test_carrier_delta_identical_to_rev1(self):
        def delta(args):
            diff = subprocess.check_output(["git", "diff", *args, "--", "README.md", GUIDE]).decode()
            return [line for line in diff.splitlines() if line.startswith(("+", "-")) and not line.startswith(("+++", "---"))]
        self.assertEqual(delta([NEW_BASE]), delta([OLD_BASE, REV1]))

    def test_refresh_preserves_both_readme_links(self):
        self.assertEqual((ROOT / "README.md").read_text(), expected_readme())

    def test_guide_bytes_identical(self):
        self.assertEqual((ROOT / GUIDE).read_bytes(), blob(REV1, GUIDE).encode())
        self.assertEqual((ROOT / GUIDE).read_bytes(), blob("ca2d4a1a", GUIDE).encode())

    def test_all_relative_links_resolve(self):
        for path in ["README.md", GUIDE]:
            links = relative_links(path, (ROOT / path).read_text())
            print(f"{path}: {len(links)}/{len(links)} relative links resolve")

    def test_readme_scope_narrowing_mutant(self):
        # A checker narrowed to the guide alone misses a broken README target.
        docs = {p: (ROOT / p).read_text() for p in ["README.md", GUIDE]}
        target = "docs/external-build-repositories.md"
        self.assertIn(f"]({target})", docs["README.md"])
        docs["README.md"] = docs["README.md"].replace(f"]({target})", "](docs/TASK-261001-1klixs-missing.md)")
        relative_links(GUIDE, docs[GUIDE])
        with self.assertRaisesRegex(AssertionError, "missing"):
            relative_links("README.md", docs["README.md"])
        print("narrowing mutant killed: guide-only scope misses broken README; full scope rejects it")

if __name__ == "__main__":
    unittest.main(verbosity=2)
