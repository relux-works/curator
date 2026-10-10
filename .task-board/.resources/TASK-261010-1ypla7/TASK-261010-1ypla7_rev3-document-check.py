"""Bounded document-only regression check; no product tests or runtime gate."""
import argparse
import hashlib
from pathlib import Path
import re

parser = argparse.ArgumentParser()
parser.add_argument('--narrow-version-options', action='store_true')
args = parser.parse_args()
p = Path('.research/261010_platform-contract-table.md')
s = p.read_text()
rows = {m.group(1): m.group(0) for m in re.finditer(r'^\| ([A-Z]\d+) .*$', s.split('## Names and filesystem', 1)[1], re.M)}
p8 = rows['P8']
if args.narrow_version_options:
    # Retain the open-status gate but narrow it to only one of the two options.
    p8 = p8.replace('(a) same version only; (b) qualified version range with a successful probe for the adapter version.', '(a) same version only.')
status = p8.rsplit(' | ', 1)[1]
checks = {
    'DB-F1 open authority and both version options': all(x in status for x in ['OWNER DECISION NEEDED', '(a) same version only', '(b) qualified version range with a successful probe for the adapter version', 'Neither is canonical', '#L413', '#L1814-L1825', '#L1973-L1977', 'must not state either rule as settled']),
    'DB-F1 unresolved value and retained preconditions': all(x in p8 for x in ['Version rule unresolved', 'same runtime family', 'same managed home', 'same profile pin', 'transcript integrity digest', 'exclusive native-session lock', 'ready authentication', 'current grants', 'no transport conversion', 'A07.5 stays open']),
    'DB-N1 two table joins': all('\n\n| '+x+' ' not in s for x in ['I5','P7']),
    'DB-N2 owner section': 'CB §§4.3/6.6 migration' in rows['W4'],
    'DB-N3 owner decision and private lanes': all(x in rows['P9'] for x in ['Decided in owner session, 2026-10-10', 'pieces of board code may appear in public CI logs', 'contract-table-rework3-brief.md', 'Lane 1 projection: private', 'Lane 2 consumer: separate private', 'does not relocate the private lanes']),
    'DB-N4 internal default and public decision': all(x in rows['P11'] for x in ['internal service-only CLI entry', 'that reading needs no new decision', 'public sweep operation would require a new owner decision']),
    'packaging': len(p.read_bytes()) <= 81920 and len(rows) == 52 and '\n## Rev3 changes\n' in s,
    'bounded hygiene': not re.search(r'/Users/|/home/|file://|https://chatgpt.com/|sk-[A-Za-z0-9]{20,}', s),
}
for name,ok in checks.items():
    print(('PASS' if ok else 'FAIL') + ': ' + name)
print(f'rev3_rejection_regression: {sum(checks.values())}/{len(checks)}; mode={"narrow-version-options" if args.narrow_version_options else "candidate"}')
print(f'candidate bytes={len(p.read_bytes())}; sha256={hashlib.sha256(p.read_bytes()).hexdigest()}')
raise SystemExit(0 if all(checks.values()) else 1)
