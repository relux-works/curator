"""Offline controls for the source and payload gates; no release keys are used."""

import importlib.util
import io
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest

spec = importlib.util.spec_from_file_location("bundle", Path(__file__).with_name("bundle.py"))
bundle = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bundle)


class SourceControls(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temporary = tempfile.TemporaryDirectory()
        cls.root = Path(cls.temporary.name)
        cls.repo = cls.root / "repo"
        cls.repo.mkdir()
        bundle.git(cls.repo, "init", "--quiet")
        bundle.git(cls.repo, "config", "user.name", "Release fixture")
        bundle.git(cls.repo, "config", "user.email", "fixture@example.invalid")
        (cls.repo / "source").write_text("fixture\n")
        bundle.git(cls.repo, "add", "source")
        bundle.git(cls.repo, "-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
        cls.commit = bundle.git(cls.repo, "rev-parse", "HEAD")[0].decode().strip()
        cls.keys = {}
        for name in ["trusted", "unknown", "historical"]:
            key = cls.root / name
            subprocess.run(["ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(key)], check=True)
            fingerprint = subprocess.check_output(["ssh-keygen", "-lf", str(key) + ".pub"]).decode().split()[1]
            cls.keys[name] = (key, fingerprint)
        cls.signers = cls.root / "allowed_signers"
        cls.signers.write_text("".join("fixture@example.invalid " + Path(str(cls.keys[name][0]) + ".pub").read_text()
                                       for name in ["trusted", "historical"]))
        cls.tag("v1.0.0", "trusted")
        cls.tag("v1.0.1", "unknown")
        cls.tag("v1.0.2", "historical")
        bundle.git(cls.repo, "-c", "tag.gpgsign=false", "tag", "-a", "v1.0.3", "-m", "unsigned")
        bundle.git(cls.repo, "-c", "tag.gpgsign=false", "tag", "v1.0.4")
        cls.pin = {"tag": "v1.0.0", "commit": cls.commit, "signer_fingerprint": cls.keys["trusted"][1]}

    @classmethod
    def tag(cls, name, signer):
        bundle.git(cls.repo, "-c", "gpg.format=ssh", "-c", "gpg.ssh.program=ssh-keygen",
                   "-c", "user.signingkey=" + str(cls.keys[signer][0]), "tag", "-s", name, "-m", name)

    @classmethod
    def tearDownClass(cls):
        cls.temporary.cleanup()

    # The positive control is a real SSH-signed tag, verified with a temporary
    # allowed-signers file. Production RELEASE_SIGNERS is never modified.
    def test_trusted_signed_tag(self):
        bundle.verify_source(self.repo, self.pin, self.signers)

    # Each rejection changes one authority condition beside the valid control.
    def test_wrong_or_missing_tag_and_commit(self):
        for delta, message in [({"tag": "v9.9.9"}, "failed"),
                               ({"commit": "0" * 40}, "commit differs")]:
            with self.subTest(delta=delta), self.assertRaisesRegex(bundle.Refusal, message):
                bundle.verify_source(self.repo, dict(self.pin, **delta), self.signers)

    def test_unknown_unsigned_and_lightweight_tags(self):
        for tag, message in [("v1.0.1", "No principal matched"), ("v1.0.3", "no signature found"), ("v1.0.4", "annotated and signed")]:
            with self.subTest(tag=tag), self.assertRaisesRegex(bundle.Refusal, message):
                bundle.verify_source(self.repo, dict(self.pin, tag=tag), self.signers)

    def test_historical_allowed_key_is_not_the_pinned_release_key(self):
        with self.assertRaisesRegex(bundle.Refusal, "pinned release key"):
            bundle.verify_source(self.repo, dict(self.pin, tag="v1.0.2"), self.signers)

    def test_renamed_signed_tag_is_rejected(self):
        obj = bundle.git(self.repo, "rev-parse", "refs/tags/v1.0.0")[0].decode().strip()
        bundle.git(self.repo, "update-ref", "refs/tags/v1.0.5", obj)
        with self.assertRaisesRegex(bundle.Refusal, "tag name differs"):
            bundle.verify_source(self.repo, dict(self.pin, tag="v1.0.5"), self.signers)

    def test_ambient_git_directory_cannot_redirect_verification(self):
        previous = os.environ.get("GIT_DIR")
        os.environ["GIT_DIR"] = str(self.root / "does-not-exist")
        try:
            bundle.verify_source(self.repo, self.pin, self.signers)
        finally:
            if previous is None:
                del os.environ["GIT_DIR"]
            else:
                os.environ["GIT_DIR"] = previous


def archive(path, launcher=True):
    with tarfile.open(path, "w:gz") as result:
        for name in ["curator", "curator-run"] if launcher else ["curator"]:
            member = tarfile.TarInfo(name)
            member.size, member.mode = 1, 0o755
            result.addfile(member, io.BytesIO(b"x"))


class PayloadControls(unittest.TestCase):
    # Nearby valid fixtures demonstrate that the gate does not reject all output.
    def test_pair_and_curator_only_windows(self):
        bundle.require_pair({"curator": (1, 0o755), "curator-run": (1, 0o755)}, "Unix")
        bundle.require_pair({"curator.exe": (1, 0), "curator-run.json": (1, 0)}, "Windows", windows=True)

    def test_missing_empty_nonexecutable_or_excluded_binary(self):
        pair = {"curator": (1, 0o755), "curator-run": (1, 0o755)}
        for change, message in [({"curator-run": (0, 0o755)}, "missing curator-run"),
                                ({"curator-run": (1, 0o644)}, "missing curator-run"),
                                ({"ax": (1, 0o755)}, "excluded ax")]:
            with self.subTest(change=change), self.assertRaisesRegex(bundle.Refusal, message):
                bundle.require_pair(dict(pair, **change), "fixture")
        with self.assertRaisesRegex(bundle.Refusal, "unexpected Windows launcher"):
            bundle.require_pair({"curator.exe": (1, 0), "curator-run.exe": (1, 0)}, "Windows", windows=True)

    def test_binary_hook_rejects_missing_empty_and_symlink(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "curator-run"
            with self.assertRaisesRegex(bundle.Refusal, "missing launcher artifact"):
                bundle.binary(path)
            path.touch()
            with self.assertRaisesRegex(bundle.Refusal, "missing launcher artifact"):
                bundle.binary(path)
            path.write_bytes(b"binary")
            bundle.binary(path)
            link = path.with_name("link")
            link.symlink_to(path)
            with self.assertRaisesRegex(bundle.Refusal, "missing launcher artifact"):
                bundle.binary(link)

    def test_missing_companion_fails_before_checksum_signer_runs(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            path = root / "curator_fixture_darwin_amd64.tar.gz"
            archive(path)
            bundle.require_pair(bundle.tar_files(path.read_bytes()), path.name)
            archive(path, launcher=False)  # Bounded plant: remove only the companion.
            signer = root / "cosign"
            marker = root / "signer-was-invoked"
            signer.write_text('#!/bin/sh\ntouch "' + str(marker) + '"\n')
            signer.chmod(0o755)
            env = dict(os.environ, PATH=str(root) + os.pathsep + os.environ["PATH"])
            result = subprocess.run(["bash", "release/sign-checksums.sh", str(root / "checksums.txt"), "unused.pem", "unused.sig"],
                                    cwd=bundle.ROOT, env=env, capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("missing curator-run", result.stderr)
            self.assertFalse(marker.exists(), "the checksum signer ran despite a missing companion")


if __name__ == "__main__":
    unittest.main()
