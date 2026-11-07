"""Reproduce the qualification evidence controls; never launch a vendor CLI."""

import argparse
import base64
import copy
import hashlib
import json
import pathlib
import sys


class EvidenceRefusal(Exception):
    pass


def selector(record, plant):
    # Proves this recorded invalid selector has its specified refusal and empty
    # stdout; the plant tests the checker, not Claude's runtime enforcement.
    data = copy.deepcopy(record)
    row = next(p for p in data["probes"] if p["argv"] == [
        "claude", "--permission-mode=invalid-mode", "--help"
    ])
    if plant:
        row["exit"] = 0
    if row["exit"] != 1:
        raise EvidenceRefusal("INVALID_SELECTOR_EXIT_MISMATCH: expected 1, observed " + str(row["exit"]))
    if data["outputs"][row["stdout_ref"]] != "":
        raise EvidenceRefusal("INVALID_SELECTOR_STDOUT_NOT_EMPTY")
    expected = "argument 'invalid-mode' is invalid. Allowed choices are acceptEdits, auto, bypassPermissions, manual, dontAsk, plan."
    if expected not in data["outputs"][row["stderr_ref"]]:
        raise EvidenceRefusal("INVALID_SELECTOR_REFUSAL_MISSING")
    print("INVALID_SELECTOR_REFUSAL_CONFIRMED")


def integrity(record, tarball, package, plant):
    # Proves bytes match the recorded published SRI. The wrong expected digest
    # must reach the intended mismatch, not fail on I/O or another setup check.
    metadata = next(p for p in record["npm_provenance"] if p["name"] == package)
    payload = tarball.read_bytes()
    if hashlib.sha256(payload).hexdigest() != metadata["tarball_sha256"]:
        raise EvidenceRefusal("TARBALL_SHA256_MISMATCH")
    expected = metadata["integrity"]
    if plant:
        prefix, digest = expected.split("-", 1)
        expected = prefix + "-" + ("A" if digest[0] != "A" else "B") + digest[1:]
    actual = "sha512-" + base64.b64encode(hashlib.sha512(payload).digest()).decode()
    if actual != expected:
        raise EvidenceRefusal("PUBLISHED_SHA512_INTEGRITY_MISMATCH")
    if hashlib.sha1(payload).hexdigest() != metadata["shasum"]:
        raise EvidenceRefusal("PUBLISHED_SHA1_MISMATCH")
    print("RECORDED_PUBLISHED_INTEGRITY_CONFIRMED")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("case", choices=["selector", "integrity"])
    parser.add_argument("record", type=pathlib.Path)
    parser.add_argument("--tarball", type=pathlib.Path)
    parser.add_argument("--package")
    parser.add_argument("--plant", action="store_true")
    args = parser.parse_args()
    record = json.loads(args.record.read_text())
    if args.case == "selector":
        selector(record, args.plant)
    else:
        if args.tarball is None or args.package is None:
            parser.error("integrity requires --tarball and --package")
        integrity(record, args.tarball, args.package, args.plant)


if __name__ == "__main__":
    try:
        main()
    except EvidenceRefusal as error:
        print(str(error), file=sys.stderr)
        sys.exit(1)
