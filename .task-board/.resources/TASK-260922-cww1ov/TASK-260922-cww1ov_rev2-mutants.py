#!/usr/bin/env python3
"""F-C3 rev2 narrowing-mutant driver (scratch, not committed).

C47 is the NEW state-only copy mutant required by the rev1 verdict R1:
it copies credential bytes ONLY into manager state (never into the
credential scope), so the rev1 row survived it. The strengthened
TestMigrateNoSecretCopies (whole-home scan + interrupted-phase scan)
must kill it. C18 (in-scope backup) is re-run to confirm the
strengthened row still kills the original shape.

One mutant at a time, tree restored after each. A killer that fails to
BUILD is INVALID, never a kill.
Usage: python3 /tmp/cww1ov_rev2_mutants.py [ID ...]  (empty = all)
"""
import subprocess
import sys

LIB = "github.com/relux-works/curator/internal/envprofile"

M = []


def mut(id, file, hunks, pkg, killer, witness, note=""):
    M.append({"id": id, "file": file, "hunks": hunks, "pkg": pkg,
              "killer": killer, "witness": witness, "note": note})


mut("C18", "internal/envprofile/migrate.go", [(
    '\t\t\tif err := atomicRelinkJournaled(req.Home, journal, i, full, op.To, req.symlinkFn(), req.renameFn()); err != nil {\n\t\t\t\treturn applied, fmt.Errorf("relink %s: %v", op.Path, err)\n\t\t\t}',
    '\t\t\tif err := atomicRelinkJournaled(req.Home, journal, i, full, op.To, req.symlinkFn(), req.renameFn()); err != nil {\n\t\t\t\treturn applied, fmt.Errorf("relink %s: %v", op.Path, err)\n\t\t\t}\n\t\t\tif payload, err := os.ReadFile(op.To); err == nil {\n\t\t\t\t_ = os.WriteFile(full+".bak", payload, 0o600)\n\t\t\t}')],
    LIB, "^TestMigrateNoSecretCopies$",
    "^TestMigrateOldRootBytesWarnAndProceed$",
    note="relink leaves a bytes backup when To exists (in-scope)")

mut("C47", "internal/envprofile/migrate.go", [(
    '\t\tcase MigrateOpRelink:\n\t\t\tif err := os.MkdirAll(filepath.Dir(full), 0o755);',
    '\t\tcase MigrateOpRelink:\n\t\t\tif payload, err := os.ReadFile(op.To); err == nil {\n\t\t\t\tif err := os.WriteFile(filepath.Join(filepath.Dir(migrationJournalPath(req.Home)), "credential-backup"), payload, 0o600); err != nil {\n\t\t\t\t\treturn applied, err\n\t\t\t\t}\n\t\t\t}\n\t\t\tif err := os.MkdirAll(filepath.Dir(full), 0o755);')],
    LIB, "^TestMigrateNoSecretCopies$",
    "^TestMigratePiWrongTargetToAgentRoot$",
    note="relink copies bytes ONLY into manager state (rev1 survivor)")


def run(pkg, mask, timeout):
    cmd = ["go", "test", pkg, "-run", mask, "-count=1",
           "-timeout=%ds" % timeout]
    p = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout + 60)
    out = p.stdout + p.stderr
    return p.returncode, out


def main():
    want = set(sys.argv[1:])
    rows = [m for m in M if not want or m["id"] in want]
    print("mutants selected: %d" % len(rows))
    fails = []
    for m in rows:
        with open(m["file"]) as f:
            orig = f.read()
        src = orig
        try:
            for anchor, repl in m["hunks"]:
                n = src.count(anchor)
                assert n == 1, "%s anchor x%d" % (m["id"], n)
                src = src.replace(anchor, repl, 1)
            with open(m["file"], "w") as f:
                f.write(src)
            krc, kout = run(m["pkg"], m["killer"], 150)
            killed = (krc != 0) and ("[build failed]" not in kout)
            invalid = (krc != 0) and ("[build failed]" in kout)
            wrc, _ = run(m["pkg"], m["witness"], 150)
            status = "INVALID-BUILD" if invalid else (
                "KILLED" if killed and wrc == 0 else
                "SURVIVED" if not killed else "WITNESS-FAILED")
            print("%s %-14s killer=%s witness=%s [%s]" % (
                m["id"], status,
                "FAIL" if krc != 0 else "pass",
                "pass" if wrc == 0 else "FAIL", m["note"]))
            if status == "KILLED":
                for line in kout.splitlines():
                    if "FAIL" in line or "credential" in line.lower():
                        print("    | " + line[:220])
            else:
                fails.append(m["id"])
                print("--- killer output ---\n" + kout[-2500:])
        finally:
            with open(m["file"], "w") as f:
                f.write(orig)
    with open("internal/envprofile/migrate.go") as f:
        assert "credential-backup" not in f.read(), "mutant residue!"
    print("done: %d/%d killed" % (len(rows) - len(fails), len(rows)))
    if fails:
        print("NEEDS-ATTENTION:", " ".join(fails))
        sys.exit(1)


main()
