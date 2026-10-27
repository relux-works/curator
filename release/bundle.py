#!/usr/bin/env python3
"""Verify the launcher source before export and release payloads before signing."""

import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import shutil
import struct
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parent.parent
PIN = ROOT / "release/curator-run.json"
SIGNERS = ROOT / "RELEASE_SIGNERS"


class Refusal(Exception):
    """A source or artifact failed the release contract."""


def pin_data(path=PIN):
    pin = json.loads(Path(path).read_text())
    if (pin["repository"] != "https://github.com/relux-works/curator-agent-launcher.git"
            or not re.fullmatch(r"v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?", pin["tag"])
            or pin["tag"] != "v" + pin["version"]
            or not re.fullmatch(r"[0-9a-f]{40}", pin["commit"])
            or not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", pin["go_version"])):
        raise Refusal("invalid launcher pin")
    return pin


def git(repo, *args):
    # Agent lanes can inherit another clone's GIT_DIR/WORK_TREE and config.
    # Verify the requested repository and public pins, never ambient Git state.
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env.update(GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_SYSTEM=os.devnull, LC_ALL="C")
    cmd = ["git", "-c", "core.hooksPath=" + os.devnull, "-C", str(repo), *args]
    result = subprocess.run(cmd, env=env, capture_output=True)
    if result.returncode:
        raise Refusal("git " + args[0] + " failed: " + result.stderr.decode(errors="replace").strip())
    return result.stdout, result.stderr


def verify_source(repo, pin, signers=SIGNERS):
    ref = "refs/tags/" + pin["tag"]
    kind = git(repo, "cat-file", "-t", ref)[0].decode().strip()
    if kind != "tag":
        raise Refusal("launcher tag must be annotated and signed")
    header = git(repo, "cat-file", "tag", ref)[0].split(b"\n\n", 1)[0]
    if b"tag " + pin["tag"].encode() not in header.splitlines():
        raise Refusal("signed tag name differs from launcher pin")
    commit = git(repo, "rev-parse", ref + "^{commit}")[0].decode().strip()
    if commit != pin["commit"]:
        raise Refusal("launcher tag commit differs from launcher pin")
    out, err = git(repo, "-c", "gpg.format=ssh", "-c", "gpg.ssh.program=ssh-keygen",
                   "-c", "gpg.ssh.allowedSignersFile=" + str(Path(signers).resolve()),
                   "verify-tag", ref)
    fingerprint = re.escape(pin["signer_fingerprint"].encode())
    if not re.search(rb"with ED25519 key " + fingerprint + rb"(?:\s|$)", out + err):
        raise Refusal("launcher tag did not use the pinned release key")


def prepare():
    pin = pin_data()
    parent = ROOT / ".tmp"
    parent.mkdir(exist_ok=True)
    destination = parent / "curator-run"
    if destination.exists() or destination.is_symlink():
        marker = destination / ".curator-bundle-source.json"
        if destination.is_symlink() or not marker.is_file():
            raise Refusal("refusing to replace unrecognized launcher staging directory")
    with tempfile.TemporaryDirectory(prefix="launcher-", dir=parent) as temporary:
        temporary = Path(temporary)
        repo = temporary / "git"
        repo.mkdir()
        git(repo, "init", "--quiet")
        git(repo, "fetch", "--no-tags", "--depth=1", pin["repository"],
            "refs/tags/" + pin["tag"] + ":refs/tags/" + pin["tag"])
        verify_source(repo, pin)
        exported = temporary / "source"
        exported.mkdir()
        archive = git(repo, "archive", "--format=tar", pin["commit"])[0]
        with tarfile.open(fileobj=io.BytesIO(archive)) as source:
            for member in source.getmembers():
                path = PurePosixPath(member.name)
                if path.is_absolute() or ".." in path.parts or not (member.isfile() or member.isdir()):
                    raise Refusal("unsupported launcher source archive entry")
            source.extractall(exported)  # Only regular files/directories beneath the verified source root.
        module = (exported / "go.mod").read_text()
        if not re.search(r"(?m)^go " + re.escape(pin["go_version"]) + r"$", module):
            raise Refusal("launcher toolchain differs from launcher pin")
        (exported / ".curator-bundle-source.json").write_text(json.dumps(pin, indent=2) + "\n")
        if destination.exists():
            shutil.rmtree(destination)  # Only this helper's marked, generated staging tree.
        exported.rename(destination)
    print("verified launcher " + pin["tag"] + " at " + pin["commit"])


def binary(path):
    path = Path(path)
    if path.is_symlink() or not path.is_file() or path.stat().st_size == 0:
        raise Refusal("missing launcher artifact: " + str(path))


def tar_files(data):
    with tarfile.open(fileobj=io.BytesIO(data)) as archive:
        return {m.name.removeprefix("./"): (m.size, m.mode)
                for m in archive.getmembers() if m.isfile()}


def deb_files(path):
    data = path.read_bytes()
    if not data.startswith(b"!<arch>\n"):
        raise Refusal("invalid deb archive")
    offset = 8
    while offset + 60 <= len(data):
        header = data[offset:offset + 60]
        size = int(header[48:58])
        name = header[:16].decode().strip().rstrip("/")
        if name.startswith("data.tar"):
            return tar_files(data[offset + 60:offset + 60 + size])
        offset += 60 + size + size % 2
    raise Refusal("deb archive lacks data payload")


def rpm_files(path):
    data = path.read_bytes()
    if data[:4] != b"\xed\xab\xee\xdb":
        raise Refusal("invalid rpm archive")
    offset = 96
    # RPM lead, signature header (8-byte aligned), then main header.
    for signature in [True, False]:
        if data[offset:offset + 3] != b"\x8e\xad\xe8":
            raise Refusal("invalid rpm header")
        count, size = struct.unpack_from(">II", data, offset + 8)
        offset += 16 + count * 16 + size
        if signature:
            offset = (offset + 7) // 8 * 8
    payload = gzip.decompress(data[offset:])  # Config explicitly selects gzip.
    result = {}
    offset = 0
    while offset + 110 <= len(payload):
        header = payload[offset:offset + 110]
        if header[:6] not in [b"070701", b"070702"]:
            raise Refusal("invalid rpm cpio payload")
        size = int(header[54:62], 16)
        name_size = int(header[94:102], 16)
        name = payload[offset + 110:offset + 110 + name_size - 1].decode()
        offset = (offset + 110 + name_size + 3) // 4 * 4
        if name == "TRAILER!!!":
            return result
        mode = int(header[14:22], 16)
        if mode & 0o170000 == 0o100000:
            result[name.removeprefix("./").lstrip("/")] = (size, mode)
        offset = (offset + size + 3) // 4 * 4
    raise Refusal("rpm cpio payload lacks trailer")


def require_pair(files, label, prefix="", windows=False):
    names = ["curator.exe"] if windows else ["curator", "curator-run"]
    for name in names:
        size, mode = files.get(prefix + name, (0, 0))
        if not size or (not windows and not mode & 0o111):
            raise Refusal("missing " + name + " in " + label)
    if any(PurePosixPath(name).name in ["ax", "ax.exe"] for name in files):
        raise Refusal("excluded ax artifact in " + label)
    if windows and any(PurePosixPath(name).name in ["curator-run", "curator-run.exe"] for name in files):
        raise Refusal("unexpected Windows launcher in " + label)


def artifacts(dist):
    dist = Path(dist)
    for goos in ["darwin", "linux", "windows"]:
        for arch in ["amd64", "arm64"]:
            extension = "zip" if goos == "windows" else "tar.gz"
            paths = list(dist.glob("curator_*_" + goos + "_" + arch + "." + extension))
            if len(paths) != 1:
                raise Refusal("missing or ambiguous release archive: " + goos + "/" + arch)
            path = paths[0]
            if goos == "windows":
                with zipfile.ZipFile(path) as archive:
                    files = {m.filename: (m.file_size, m.external_attr >> 16) for m in archive.infolist()}
            else:
                files = tar_files(path.read_bytes())
            require_pair(files, path.name, windows=goos == "windows")
    for extension, reader in [("deb", deb_files), ("rpm", rpm_files)]:
        for arch in ["amd64", "arm64"]:
            paths = list(dist.glob("curator_*_linux_" + arch + "." + extension))
            if len(paths) != 1:
                raise Refusal("missing or ambiguous " + extension + ": " + arch)
            require_pair(reader(paths[0]), paths[0].name, prefix="usr/bin/")
    print("verified six archives and four Linux packages; Unix pair, Windows curator only")


def channels(dist):
    """Check generated channel manifests against the actual archive digests."""
    dist = Path(dist)
    sums = {}
    for line in (dist / "checksums.txt").read_text().splitlines():
        digest, name = line.split(maxsplit=1)
        sums[name.strip()] = digest
    archives = sorted([*dist.glob("curator_*.tar.gz"), *dist.glob("curator_*.zip")])
    if len(archives) != 6:
        raise Refusal("channel verification requires six release archives")
    manifests = {}
    for channel, filename in [("darwin", "curator.rb"), ("windows", "curator.json"), ("linux", "curator-bin.pkgbuild")]:
        paths = list(dist.rglob(filename))
        if len(paths) != 1:
            raise Refusal("missing or ambiguous channel manifest: " + filename)
        manifests[channel] = paths[0].read_text()
    cask = manifests["darwin"]
    if not all(re.search(r'(?m)^\s*binary [\'"]' + name + r'[\'"]', cask)
               for name in ["curator", "curator-run"]):
        raise Refusal("cask must install both binaries")
    scoop = json.loads(manifests["windows"])
    architectures = scoop.get("architecture", {})
    if set(architectures) != {"64bit", "arm64"} or any(
            value.get("bin") != ["curator.exe"] for value in architectures.values()):
        raise Refusal("Scoop must install Curator only on both Windows architectures")
    aur = manifests["linux"]
    if 'install -Dm755 curator-run "${pkgdir}/usr/bin/curator-run"' not in aur:
        raise Refusal("AUR must install the launcher beside curator")
    if 'install -Dm644 release/curator-run.json ' not in aur:
        raise Refusal("AUR must install the archived source pin")
    srcinfo = dist / "aur/curator-bin.srcinfo"
    if not srcinfo.is_file():
        raise Refusal("missing AUR .SRCINFO")
    for path in archives:
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        platform = next(name for name in manifests if "_" + name + "_" in path.name)
        if sums.get(path.name) != digest or digest not in manifests[platform]:
            raise Refusal("channel checksum differs from archive: " + path.name)
        if platform == "linux":
            payload = tar_files(path.read_bytes())
            for source in re.findall(r"(?m)^\s*install -Dm[0-7]+ (\S+) ", aur):
                if source not in payload:
                    raise Refusal("AUR install source absent from archive: " + source)
        sbom = dist / (path.name + ".sbom.json")
        if sums.get(sbom.name) != hashlib.sha256(sbom.read_bytes()).hexdigest():
            raise Refusal("SBOM checksum differs: " + sbom.name)
        if platform == "linux" and digest not in srcinfo.read_text():
            raise Refusal("AUR .SRCINFO checksum differs: " + path.name)
        document = json.loads(sbom.read_text())
        if document.get("spdxVersion") != "SPDX-2.3" or not document.get("packages"):
            raise Refusal("missing or empty SPDX archive SBOM: " + path.name)
    for path in [*dist.glob("curator_*.deb"), *dist.glob("curator_*.rpm")]:
        if sums.get(path.name) != hashlib.sha256(path.read_bytes()).hexdigest():
            raise Refusal("package checksum differs: " + path.name)
    print("verified cask, Scoop and AUR manifests, archive/package checksums and six SBOMs")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=["prepare", "binary", "artifacts", "channels", "credentials"])
    parser.add_argument("path", nargs="?")
    args = parser.parse_args()
    if args.command in ["binary", "artifacts", "channels"] and args.path is None:
        parser.error(args.command + " requires a path")
    try:
        if args.command == "prepare":
            prepare()
        elif args.command == "binary":
            binary(args.path)
        elif args.command == "artifacts":
            artifacts(args.path)
        elif args.command == "channels":
            channels(args.path)
        else:
            names = ["MACOS_SIGN_P12", "MACOS_SIGN_P12_PASSWORD", "ASC_ISSUER_ID", "ASC_KEY_ID", "ASC_KEY_P8", "AUR_SSH_PRIVATE_KEY", "TAP_GITHUB_TOKEN"]
            missing = [name for name in names if not os.environ.get(name)]
            if missing:
                raise Refusal("release credentials not configured: " + ", ".join(missing))
    except (Refusal, OSError, ValueError, KeyError, tarfile.TarError, zipfile.BadZipFile, struct.error) as error:
        parser.exit(1, "bundle: " + str(error) + "\n")


if __name__ == "__main__":
    main()
