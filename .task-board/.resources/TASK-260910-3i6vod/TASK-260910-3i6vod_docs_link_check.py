import argparse
import pathlib
import re
import subprocess
import sys
from urllib.parse import urlparse

PIN = "23435129ebc4c29e5b7f75ec72a0aa0cd3f16065"
SPEC = pathlib.Path("/Users/administrator/Developer/ReluxWorks/curator/curator-spec")


def fail(message):
    print(f"FAIL: {message}")
    return False


def section(text, heading):
    lines = text.splitlines()
    try:
        start = lines.index(heading)
    except ValueError:
        return None
    body = []
    for line in lines[start + 1:]:
        if line.startswith("## "):
            break
        body.append(line)
    return "\n".join(body)


def slug(heading):
    value = re.sub(r"^#+\s*", "", heading).strip().lower().replace(chr(96), "")
    value = re.sub(r"[^\w\s-]", "", value)
    return re.sub(r"[\s-]+", "-", value).strip("-")


def spec_markdown(path):
    result = subprocess.run(
        ["git", "-C", str(SPEC), "show", f"{PIN}:{path}"],
        text=True,
        capture_output=True,
    )
    if result.returncode:
        raise RuntimeError(f"pinned spec file missing: {path}")
    return result.stdout


def anchor_exists(markdown, wanted):
    return any(slug(line) == wanted for line in markdown.splitlines() if line.startswith("#"))


def check_links(text, name, include_local):
    ok = True
    for label, target in re.findall(r"\[([^\]]+)\]\(([^)]+)\)", text):
        parsed = urlparse(target)
        if parsed.scheme in ("http", "https"):
            match = re.fullmatch(r"/relux-works/curator-spec/blob/([^/]+)/(.+)", parsed.path)
            if parsed.netloc != "github.com" or not match or match.group(1) != PIN:
                ok = fail(f"{name}: unexpected spec link {target}") and ok
                continue
            path, anchor = match.group(2), parsed.fragment
            markdown = spec_markdown(path)
            if not anchor or not anchor_exists(markdown, anchor):
                ok = fail(f"{name}: missing pinned spec anchor {target}") and ok
        elif include_local and target.startswith("SECURITY.md#"):
            local_path, anchor = target.split("#", 1)
            target_doc = pathlib.Path(args.readme).parent / local_path
            if not target_doc.is_file() or not anchor_exists(target_doc.read_text(), anchor):
                ok = fail(f"{name}: unresolved local link {target}") and ok
    return ok


def check_doc(path, heading, name, include_local):
    doc = pathlib.Path(path).read_text()
    body = section(doc, heading)
    if body is None:
        return fail(f"{name}: missing section {heading}")
    lowered = re.sub(r"\s+", " ", body.lower())
    required = [
        "run as the invoking user",
        "operating-system privileges",
        "portable assurance",
        "script-worker-v1",
        "declared-only",
        "verified",
        "provider-backed enforcement path",
    ]
    ok = True
    for phrase in required:
        if phrase not in lowered:
            ok = fail(f"{name}: missing required statement {phrase!r}") and ok
    paragraphs = re.split(r"\n\s*\n", body.lower())
    privilege_posture = any(
        "run as the invoking user" in re.sub(r"\s+", " ", paragraph)
        and "operating-system privileges" in re.sub(r"\s+", " ", paragraph)
        and "portable assurance" in re.sub(r"\s+", " ", paragraph)
        for paragraph in paragraphs
    )
    if not privilege_posture:
        ok = fail(f"{name}: user-privilege statement is not tied to portable assurance") and ok
    sentences = re.split(r"(?<=[.!?])\s+", lowered)
    script_enforcement = any(
        "script-worker-v1" in sentence and "enforced" in sentence
        for sentence in sentences
    )
    if not script_enforcement:
        ok = fail(f"{name}: script-worker-v1 is not identified as an enforced path") and ok
    verified_enforcement = any(
        "verified" in sentence and "provider-backed enforcement path" in sentence
        for sentence in sentences
    )
    if not verified_enforcement:
        ok = fail(f"{name}: verified mode is not identified as a provider-backed enforcement path") and ok
    for anchor in (
        "#411-portable-script-worker-v1-execution-policy",
        "#421-portable-manager-worker-v1-execution-policy",
        "#1-closed-selection",
        "#2-platform-neutral-provider-contract",
        "#5-failure-rules",
    ):
        if anchor not in body:
            ok = fail(f"{name}: missing curator-spec link anchor {anchor}") and ok
    return check_links(body, name, include_local) and ok


parser = argparse.ArgumentParser()
parser.add_argument("--readme", default="README.md")
parser.add_argument("--security", default="SECURITY.md")
parser.add_argument("--spec", default=str(SPEC))
args = parser.parse_args()
SPEC = pathlib.Path(args.spec)
readme_ok = check_doc(args.readme, "## Installed command security", "README.md", True)
security_ok = check_doc(args.security, "## Installed command execution", "SECURITY.md", False)
if readme_ok and security_ok:
    print("PASS: installed-command security wording and pinned links resolve")
    sys.exit(0)
sys.exit(1)
