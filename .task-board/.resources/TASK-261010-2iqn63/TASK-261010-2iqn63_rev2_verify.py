"""Document-only rev2 checks. No product code, tests, builds, or mutations.
Usage: python3 rev2_verify.py citations|regression REPORT.md
Citations fetch pinned GitHub blobs through gh; regression checks those audited locators.
"""
import base64
import concurrent.futures
import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path

ROWS = {'A02.2', 'A04.4', 'A04.5', 'A06.1', 'A06.2', 'A07.6', 'A08.1', 'A08.3', 'A08.4', 'A09.4', 'A09.9'} | {f'A11.{n}' for n in range(1, 12)}
LINK = re.compile(r'\[([^\]]+)\]\(https://github\.com/relux-works/([^/]+)/blob/([0-9a-f]{40})/([^ #]+)#L(\d+)(?:-L(\d+))?(?: "[^"]*")?\)')

def rows(text):
    text = text.split('## A.', 1)[1].split('## B.', 1)[0]
    return {m.group(1): m.group(0) for m in re.finditer(r'^\| (A\d\d\.\d+) \|.*$', text, re.M)}

def citations(text):
    selected = rows(text)
    assert ROWS <= selected.keys(), 'missing requested audit rows'
    links = [m for rid in sorted(ROWS) for m in LINK.finditer(selected[rid])]
    unique = {(m[2], m[3], m[4]) for m in links}
    def fetch(key):
        repo, commit, path = key
        p = subprocess.run(['gh', 'api', f'repos/relux-works/{repo}/contents/{path}?ref={commit}'], capture_output=True)
        assert p.returncode == 0, f'gh read failed with exit {p.returncode}: {repo}/{path}'
        data = json.loads(p.stdout)
        assert data['encoding'] == 'base64', 'unsupported content encoding'
        blob = base64.b64decode(data['content'])
        assert hashlib.sha1(b'blob '+str(len(blob)).encode()+b'\0'+blob).hexdigest() == data['sha'], 'blob integrity failure'
        return key, blob.decode().splitlines()
    with concurrent.futures.ThreadPoolExecutor(max_workers=6) as pool:
        content = dict(pool.map(fetch, sorted(unique)))
    for m in links:
        lo, hi = int(m[5]), int(m[6] or m[5])
        assert 1 <= lo <= hi <= len(content[(m[2], m[3], m[4])]), f'bad range: {m[1]}'
        assert m[1].endswith(f'L{lo}' + (f'–{hi}' if hi != lo else '')), f'label-anchor mismatch: {m[1]}'
    c09 = content[('curator-spec', '1604d4022f97d6f0751e17a37c549d2a10581829', 'cips/CIP-0009-donor-side-deployment-and-bridge.md')]
    for n, phrase in [(145,'What the worker sees'),(149,'D7.'),(160,'D9.'),(162,'D10.'),(164,'D11.'),(169,'first slice')]:
        assert phrase in c09[n-1], f'CIP decision locator drift at L{n}'
    print(json.dumps({'check':'rev2_changed_row_citations','requested_rows':f'{len(ROWS)}/{len(ROWS)}','citation_ranges':f'{len(links)}/{len(links)}','pinned_gh_reads_exit_0':f'{len(unique)}/{len(unique)}','blob_hashes_verified':len(unique),'bound':'Range, label and named CIP block checks; semantic judgements require human review.'}))

def regression(text):
    r = rows(text)
    # Read submitted report rows themselves: summary/changelog occurrences cannot satisfy checks.
    row = r['A09.9']
    for phrase in ['Capacity shortage waits', 'quota excess permanently refuses', '**v0 refusal**', '**v1 queueing**', 'quota refusal does not queue']:
        assert phrase in row, 'A09.9 capacity/quota/phase regression'
    requirements = {'A06.1': [145, 164, 169], 'A08.4': [145, 164, 169], 'A11.7': [160], 'A11.8': [162], 'A11.6':[128,133,144,145], 'A11.10':[139,145,164,169], 'A11.11':[149,154]}
    for rid, targets in requirements.items():
        coverage = set()
        for m in LINK.finditer(r[rid]):
            if m[2] == 'curator-spec' and m[4].endswith('CIP-0009-donor-side-deployment-and-bridge.md'):
                assert m[3] == '1604d4022f97d6f0751e17a37c549d2a10581829', 'wrong CIP pin'
                coverage.update(range(int(m[5]), int(m[6] or m[5])+1))
        assert set(targets) <= coverage, f'{rid} missing defining CIP lines {sorted(set(targets)-coverage)}'
    for rid in ['A02.2','A07.6','A08.3','A09.4']:
        assert 'CONFLICTING' not in r[rid], f'{rid} overstates conflict'
    for rid in ['A06.2','A08.1']:
        assert 'agent-session-bridge' in r[rid] and 'design stub' in r[rid], f'{rid} bridge stub omitted'
    for rid in ['R-LR1','R-CS2']:
        line = next(l for l in text.splitlines() if l.startswith('| P0 / '+rid+' |'))
        assert 'OWNER DECISION' in line, f'{rid} owner decision omitted'
    assert text.index('| P0 / R-CS2 |') < text.index('| P0 / R-CB1 |'), 'dependency ordering'
    for rid in ['R-SH1','R-CB1','B-F1']:
        line = next(l for l in text.splitlines() if l.startswith('| P0 / '+rid+' |'))
        assert not re.search(r'A\d\d\.\d+[–-]A?\d',line), f'{rid} uses broad row ranges'
    assert '"wiki/session-host/architecture.ru.md — 7.8 ' not in text, 'private heading retained'
    assert text.count('"wiki/session-host/architecture.ru.md — §7.8"') == 5, 'neutral link titles'
    print(json.dumps({'check':'rev2_rejection_regression','result':'PASS','scope':'Rejected phase, CIP locators, qualifier, owner-decision, dependency, row-ID and hygiene findings; documentary assertions only.'}))

if __name__ == '__main__':
    try:
        mode, filename = sys.argv[1:]
        {'citations':citations,'regression':regression}[mode](Path(filename).read_text())
    except (AssertionError, ValueError, KeyError) as error:
        print(f'FAIL: {error}')
        sys.exit(1)
