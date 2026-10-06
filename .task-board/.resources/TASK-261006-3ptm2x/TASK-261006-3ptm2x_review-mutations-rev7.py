import subprocess, pathlib, json, re, time, sys, os
root=pathlib.Path('/tmp/opaque-review-rev7')
log=pathlib.Path('/tmp/TASK-261006-3ptm2x_review-checks-rev7.jsonl')
template=pathlib.Path('/tmp/opaque-empty-git-template'); template.mkdir(exist_ok=True)
os.environ['GIT_TEMPLATE_DIR']=str(template)

def clean(s):
    s=re.sub(r'Document:\[[0-9 ]*\]', 'Document:<launch-fragment bytes>', s)
    s=s.replace(str(pathlib.Path.home()), '<operator-home>').replace(str(root), '<disposable-copy>')
    s=re.sub(r'/private/var/folders/[^\s:)]+', '<temporary-path>', s)
    s=re.sub(r'/var/folders/[^\s:)]+', '<temporary-path>', s)
    s=re.sub(r'/tmp/go-build[^\s:)]+', '<go-work>', s)
    return s

def health():
    p=subprocess.run(['launchctl','print','system/com.apple.security.syspolicy'],text=True,capture_output=True)
    count=re.search(r'successive crashes = (\d+)',p.stdout)
    return {'running':bool(re.search(r'^\s*state = running$',p.stdout,re.M)), 'crashes':int(count[1]) if count else None}

def run(label,pkg,mask,expect=0):
    before=health()
    if not before['running']: raise RuntimeError('syspolicyd not running; do not start build')
    cmd=[str(pathlib.Path.home()/'.local/bin/mini-build-lock'),'run','opaque','--','env','GOFLAGS=-work','go','test','./internal/'+pkg,'-run',mask,'-count=1','-timeout=6m','-v']
    started=time.monotonic()
    p=subprocess.run(cmd,cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=480)
    after=health()
    row={'label':label,'package':pkg,'mask':mask,'expected_exit':expect,'exit':p.returncode,'duration_seconds':round(time.monotonic()-started,3),'health_before':before,'health_after':after,'output':clean(p.stdout)}
    with log.open('a') as f: f.write(json.dumps(row)+'\n')
    print(json.dumps({k:v for k,v in row.items() if k!='output'}))
    print(clean('\n'.join(x for x in p.stdout.splitlines() if 'FAIL' in x or 'want' in x or 'computed' in x or 'refusal' in x or x.startswith('ok\t'))))
    if before != after: raise RuntimeError('host state changed: wait five minutes before further builds')
    if p.returncode != expect: raise RuntimeError('unexpected test outcome')
    if expect == 1 and ('[build failed]' in p.stdout or '--- FAIL:' not in p.stdout): raise RuntimeError('invalid mutant; no behavioral test failure')

def edit_once(path,old,new):
    target=root/path
    content=target.read_text()
    assert content.count(old)==1, (path,content.count(old),old[:80])
    target.write_text(content.replace(old,new))

baselines={
'audit':r'^Test(Gate(V1|V2|Mixed|Rejects)|AuditSubject(V1|WithOpaque)|CheckSourceAudit(V1|NUL)|V[12](Rejects|Ignores|Honors|Verdict)|Pin(AtVersion|VersionMatch)|LegacySchemaless)',
'envprofile':r'^Test(Resolve(RefusesV1CommitPinnedContextNUL|CleanV1CommitContextStaysCurrent|AdmitsV2CommitPinnedContextNUL)|SwitchMaterializeRefusesV1CommitPinnedContextNUL|GuardedReadersV1NULRefusalComputesNoV1Identity|StrictAuditMemberRefusesV1NULBeforeHashing|ValidateNamedStorePinsRefusesV1NULBeforeHashing|SkillsOfRefusesV1NULBeforeHashing)$',
'contextaudit':r'^Test(DetectAtVersionAppliesNULRuleOnlyToV1|DetectFiles|DetectBlocks|Opaque)',
'opaquescan':'.',
'hashing':r'V1|V2|CountV1|ContentSHA256',
}
if sys.argv[1]=='baseline':
    pkg=sys.argv[2]; run('baseline-'+pkg,pkg,baselines[pkg]); sys.exit()
if sys.argv[1]=='probes':
    run('independent-production-probes','audit','^TestReview(Frozen|V2Rejects|NewV2|ResolveDraft|LegacyMarker|SourceAudit)'); sys.exit()

# All mutants are confined to the disposable copy and byte-restored after each call.
name=sys.argv[1]
paths=['internal/envprofile/store_boundary.go','internal/envprofile/switch.go','internal/audit/audit.go','internal/contextaudit/contextaudit.go','internal/closure/resolve.go','internal/marker/marker.go']
original={p:(root/p).read_bytes() for p in paths}
try:
    if name in ['commit-pin','commit-both']:
        guard='\tif err := opaquescan.RefuseNULV1(entry, version); err != nil {\n\t\treturn "", "", err\n\t}\n'
        edit_once('internal/envprofile/store_boundary.go',guard, '')
        edit_once('internal/envprofile/store_boundary.go','\tactual, err = hashing.ContentSHA256WithVersion(entry, map[string]bool{}, version)',guard+'\tactual, err = hashing.ContentSHA256WithVersion(entry, map[string]bool{}, version)')
    if name in ['materialization','commit-both']:
        edit_once('internal/envprofile/switch.go', '\t\"github.com/relux-works/curator/internal/opaquescan\"\n', '')
        edit_once('internal/envprofile/switch.go','\t\tif err := opaquescan.RefuseNULV1(entry, lock.ContentHashVersion()); err != nil {\n\t\t\treturn nil, fmt.Errorf("member %s: %v", member.Name, err)\n\t\t}\n','')
    if name in ['commit-pin','commit-both']:
        run(name,'envprofile','^TestResolveRefusesV1CommitPinnedContextNUL$',1)
    elif name=='materialization':
        run(name,'envprofile','^TestSwitchMaterializeRefusesV1CommitPinnedContextNUL$',1)
    elif name=='hash-before-refuse':
        edit_once('internal/audit/audit.go','\t\topaqueReport := opaqueNULReport(opaquePaths)','\t\t_, _ = hashing.ContentSHA256WithVersion(subject.Snapshot, nil, hashing.VersionV1)\n\t\topaqueReport := opaqueNULReport(opaquePaths)')
        run(name,'audit','^Test(AuditSubject|CheckSourceAudit)V1NULRefusalComputesNoV1Identity$',1)
    elif name=='gate-hash-before-refuse':
        edit_once('internal/audit/audit.go','\t\treport := opaqueNULReport(paths)','\t\t_, _ = hashing.ContentSHA256WithVersion(subject.Snapshot, nil, hashing.VersionV1)\n\t\treport := opaqueNULReport(paths)')
        run(name,'audit','^TestGateV1NULRefusalComputesNoV1Identity$',1)
    elif name=='pin-version':
        edit_once('internal/audit/audit.go','\treturn recordVersion == want','\treturn recordVersion == want || recordVersion == hashing.VersionV1 || recordVersion == hashing.VersionV2')
        run(name,'audit','^Test(V2RejectsLegacyV1Pin|V1RejectsV2Pin|LegacySchemalessPinReadsAsV1)$',1)
    elif name=='cache-version':
        edit_once('internal/audit/audit.go','\tif recordVersion != version {','\tif recordVersion != version && false {')
        run(name,'audit','^Test(V2RejectsLegacyUnversionedVerdict|V1IgnoresV2VerdictCarrier)$',1)
    elif name=='cache-v1-carrier':
        edit_once('internal/audit/audit.go','\t\trecord["schema_version"] = verdictSchemaV2\n\t\trecord["hash_version"] = int(version)','\t\trecord["schema_version"] = verdictSchemaV1')
        run(name,'audit','^TestV2VerdictCarrierRecordsHashVersion$',1)
    elif name=='v2-blocked':
        edit_once('internal/audit/audit.go','\t\tif version != hashing.VersionV1 {','\t\tif version != hashing.VersionV1 && version != hashing.VersionV2 {')
        run(name,'audit','^TestGateV2AdmitsNULBearingTreeWhenAuditIsDisabled$',1)
    elif name=='legacy-as-v2':
        edit_once('internal/audit/audit.go','\tcase 0, hashing.VersionV1:\n\t\treturn hashing.VersionV1, nil','\tcase 0:\n\t\treturn hashing.VersionV2, nil\n\tcase hashing.VersionV1:\n\t\treturn hashing.VersionV1, nil')
        run(name,'audit','^TestGateV1KeepsOpaqueBlockWithUnchangedFinding$',1)
    elif name=='context-v2-blocked':
        edit_once('internal/contextaudit/contextaudit.go','\tif hashVersion == hashing.VersionV1 {\n\t\treturn Detect(root, pin, waivers)','\tif hashVersion == hashing.VersionV1 || hashVersion == hashing.VersionV2 {\n\t\treturn Detect(root, pin, waivers)')
        run(name,'contextaudit','^TestDetectAtVersionAppliesNULRuleOnlyToV1$',1)
    elif name=='frozen-projection':
        guard='\tif err := opaquescan.RefuseNULV1(frozen, hashing.VersionV1); err != nil {\n\t\treturn "", fmt.Errorf("source_member_invalid: frozen package context: %v", err)\n\t}\n'
        edit_once('internal/closure/resolve.go',guard,'')
        edit_once('internal/closure/resolve.go','\tdigest, err := hashing.ContentSHA256(destination, nil)',guard.replace('(frozen,','(destination,')+'\tdigest, err := hashing.ContentSHA256(destination, nil)')
        run(name,'audit','^TestReview(FrozenV1RefusesExcludedNUL|ResolveDraftRefusesExcludedNUL)$',1)
    elif name=='marker-legacy-narrowed':
        edit_once('internal/marker/marker.go','\tif err := opaquescan.RefuseNULV1(installedDir, recordedHashVersion); err != nil {','\tif err := opaquescan.RefuseNULV1(installedDir, hashing.VersionV2); err != nil {')
        run(name,'audit','^TestReviewLegacyMarkerCollisionAndDowngrade$',1)
    else: raise ValueError(name)
finally:
    for p,data in original.items(): (root/p).write_bytes(data)
    assert all((root/p).read_bytes()==data for p,data in original.items())
