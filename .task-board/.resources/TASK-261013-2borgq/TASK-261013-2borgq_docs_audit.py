from pathlib import Path
import re, subprocess
files = subprocess.check_output(['git','diff','--name-only'],text=True).splitlines()
allowed = {'SPEC.md','README.md','CHANGELOG.md','diagrams/README.md',
           'diagrams/components.puml','diagrams/launch-paths.puml',
           'diagrams/launch-states.puml','diagrams/artifact-layout.puml'}
assert set(files) == allowed, files
print(f'Docs-only scope: {len(files)}/{len(allowed)} expected files; no code/schema/logbook changes')
diff = subprocess.check_output(['git','diff','--unified=0'],text=True)
added = '\n'.join(l[1:] for l in diff.splitlines() if l.startswith('+') and not l.startswith('+++'))
# Audit additions only. Existing historical identifiers and UTF-8 remain untouched.
for pattern in (r'(?<![A-Za-z0-9])F-\d+', r'\bf-(?:wire-register-drift|book-muse-c7|exec-input-closure|exec-profile-authority|base-binding|workspace-result-modes)\b', r'/Users/',r'/home/',r'board-resource',r'app://',r'\$[A-Za-z0-9_-]{25,}'):
    assert not re.search(pattern,added), pattern
print('Public added-text hygiene: 7/7 scoped patterns absent; no claim of a universal name classifier')
spec=Path('SPEC.md').read_text()
checks=[
 ('manifest from immutable objects', "derived from B's immutable repository objects" in spec),
 ('materialization exact entries', 'any mismatch, missing or extra entry MUST' in spec),
 ('independent repository import', "recompute the canonical manifest from the caller repository's immutable B" in spec),
 ('result-to-B equality', "the result's `base_commit` equals B" in spec),
 ('result-to-manifest equality', "result's `manifest_digest`" in spec),
 ('mode-aware export', 'compares both kind/mode and content' in spec),
 ('mode-only and type transitions', 'Mode-only edits' in spec and 'equal-blob transitions' in spec),
 ('deletion prior mode', "deletions report the manifest's prior mode" in spec),
 ('protected selector', 'policy, beside R-NETBIND, maps' in spec),
 ('profile-body pin', 'resolved identity and body digest in `input/1`' in spec),
 ('pre-child verification', 'Before any child code starts' in spec),
 ('full invocation and inputs', 'complete admitted invocation' in spec and 'every transitive input' in spec),
 ('input refusal qualification', 'scripts, inline evaluation, imported modules, project plugins/executable' in spec),
 ('inert controls and off-host fallback', 'positive controls for genuinely inert light checks' in spec and 'operation sent to R-VERIFY' in spec),
 ('Muse three states', 'Priority decided 2026-10-13' in spec and 'channel is\nproposed and qualification remains open' in spec),
 ('owner references', 'Owner-linked format reference index' in spec and 'blob/74687e600e04817fa6e7a2e8904a9f7a3c83a1d8' in spec),
 ('draft/transport distinction', 'draft result fields' in spec and 'blob delivery, retention and activation remain gates' in spec),
 ('no stale scoped status', not any(x in spec[spec.index('## 10.'): ] for x in ['path to blob id', 'schema and retention remain\nunpublished', 'dispatcher v0', 'subscription qualification and choose a channel'])),
 ('release heading', '## Unreleased\n\n## 0.3.0 — 2026-10-13' in Path('CHANGELOG.md').read_text()),
 ('current spec and dependency references', '[SPEC 0.5.1-draft]' in Path('README.md').read_text() and 'v0.5.48' not in spec),
]
for name, ok in checks:
    assert ok, name
print(f'Document consistency clauses: {len(checks)}/{len(checks)} present; text audit only, no implementation qualification')
count=0
for fn in files:
    p=Path(fn);s=p.read_text()
    links=re.findall(r'\]\(([^)]+)\)',s) if p.suffix=='.md' else re.findall(r'\[\[([^\s]+)',s)
    for link in links:
        if '://' in link or link.startswith('#'): continue
        target=link.split('#')[0]
        assert (p.parent/target).is_file(), (fn,link)
        count+=1
print(f'Local link file targets: {count}/{count} exist; external HTTP availability and anchors not checked')
for p in sorted(Path('diagrams').glob('*.puml')):
    lines=p.read_text().splitlines();in_note=False;groups=0;seq=[]
    assert lines[0]=='@startuml' and lines[-1]=='@enduml',p
    for number,line in enumerate(lines,1):
        t=line.strip()
        if in_note:
            if t=='end note': in_note=False
            continue
        if t.startswith('note '):
            in_note=True;continue
        if not t: continue
        if t.endswith('{'): groups+=1
        if t=='}': groups-=1;assert groups>=0,(p,number)
        if p.name=='launch-paths.puml':
            if t.startswith(('alt ','opt ','loop ','par ','group ')): seq.append(t)
            elif t=='end': assert seq,(p,number);seq.pop()
            elif t.startswith('else '): assert seq and seq[-1].startswith('alt '),(p,number)
            known=t.startswith(('@startuml','@enduml','title ','actor ','participant ','== ','alt ','opt ','loop ','par ','group ','else ')) or t=='end' or ' ->' in t or ' -->' in t
        else:
            known=t.startswith(('@startuml','@enduml','title ','skinparam ','package ','actor ','component ','rectangle ','database ','folder ','artifact ','state ')) or t=='}' or ' -->' in t
        assert known,(p,number,t)
    assert not in_note and groups==0 and not seq,(p,groups,seq)
    s=p.read_text()
    assert 'session_host_missing' in s and 'v0.3.0' in s and 'until swarma dispatcher path ships' in s,p
    assert 'hosted Phase 1' not in s and 'schema unpublished' not in s and 'dispatcher v0' not in s,p
    assert 'mode-only' in s and 'body digest' in s and 'transitive' in s,p
print('Figure source structure/obligations: 4/4 pass; limited lexical/block checks, NOT PlantUML parsing/rendering')
print('Tests/builds: not run, expressly excluded by the docs-only brief')
print('Rendering: not run; java -version exited 1 (no runtime); PlantUML absent')
