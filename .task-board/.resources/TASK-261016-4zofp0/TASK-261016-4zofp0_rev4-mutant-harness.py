import pathlib, subprocess, shutil, json
root=pathlib.Path.cwd()
base=root/'.temp/SNAPLOCK-rev4/mutant-workspace'
base.mkdir(parents=True,exist_ok=True)
files=subprocess.check_output(['git','ls-files','--cached','--others','--exclude-standard','-z']).decode().split('\0')
for name in filter(None,files):
 if name.split("/",1)[0] not in {"internal","cmd","tools","scripts",".spec","go.mod","go.sum","README.md"}:continue
 src=root/name;dst=base/name
 if not src.is_file():continue
 dst.parent.mkdir(parents=True,exist_ok=True)
 if src.is_symlink():
  if dst.exists() or dst.is_symlink():dst.unlink()
  dst.symlink_to(src.readlink())
 else:shutil.copy2(src,dst)
mutants=[
 ('M-policy-probe','internal/envprofile/credential_policy.go','if req.credentialRelease != nil {','if req.credentialRelease != nil && adapter.ID != envregistry.CodexCLI {','./internal/envprofile','^TestMaterializeCaptureAndStatusNeverProbeVendor$'),
 ('M-release-cache','internal/envprofile/credential_policy.go','cached.release == release','(cached.release == release || release == "unknown")','./internal/envprofile','^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$'),
 ('M-codex-file-admission','internal/envprofile/credential_policy.go','if selection == envregistry.CredentialStoreNative {','if selection == envregistry.CredentialStoreNative || (adapter.ID == envregistry.CodexCLI && selection == envregistry.CredentialStoreFile) {','./internal/envprofile','^TestMaterializeCaptureCredentialStoreAdmission$'),
 ('M-denied-class','internal/transaction/pending.go','return false, stateread.UnusableError(pathErr.Path, err)','if errors.Is(err, os.ErrPermission) { return false, err }; return false, stateread.UnusableError(pathErr.Path, err)','./internal/envprofile','^TestMaterializeCaptureUnreadableJournal$'),
 ('M-lock-name','internal/managerlock/managerlock.go','!named.Mode().IsRegular() || !os.SameFile(held, named)','!named.Mode().IsRegular() || (!os.SameFile(held, named) && named.Size() != 0)','./internal/envprofile','^TestMaterializeCaptureAuthorizingDrift/lock-identity$'),
 ('M-composition-policy','internal/envprofile/materialize_capture.go','if err != nil || !bytes.Equal(baseline, nextPolicy) {','if err != nil || (!bytes.Equal(baseline, nextPolicy) && !func() bool { permitted := selected.Policy; permitted.OverlaysAllowed = true; permittedBytes, e := json.Marshal(struct { Machine envregistry.MachineConfig; Policy Policy }{selected.Machine, permitted}); return e == nil && bytes.Equal(permittedBytes, nextPolicy) }()) {','./internal/envprofile','^TestMaterializeCaptureSameFragmentDrift/composition-policy$'),
 ('M3-lock-only','internal/envprofile/materialize_capture.go','for _, object := range c.objects {\n\t\tif c.ctx.Err()', 'for _, object := range c.objects[1:2] {\n\t\tif c.ctx.Err()', './internal/envprofile','^TestMaterializeCaptureSameFragmentDrift/aba$'),
 ('M-prepared','internal/transaction/pending.go','return len(ids) != 0, nil','for _, id := range ids { j, err := engine.loadJournal(id); if err != nil { return false, err }; if j.Phase != PhasePrepared { return true, nil } }; return false, nil','./internal/envprofile','^TestMaterializeCapturePendingDoesNotRecover/prepared$'),
 ('M-corrupt-absence','internal/transaction/pending.go','if _, err := engine.loadJournal(id); err != nil {','if _, err := engine.loadJournal(id); err != nil { if id == "broken" { return false, nil }','./internal/envprofile','^TestMaterializeCaptureBadRecoveryEvidence/corrupt$'),
 ('M-probe','internal/envprofile/managed.go','return verifyHomeObserved(req, adapter, source, lock, hash, "unknown")','if adapter.ID == envregistry.CodexCLI { return verifyHomeObserved(req, adapter, source, lock, hash, req.detected(adapter)) }; return verifyHomeObserved(req, adapter, source, lock, hash, "unknown")','./internal/envprofile','^TestMaterializeCaptureAndStatusNeverProbeVendor$'),
 ('M-executable-mode','internal/envprofile/materialize_capture.go','a.Mode() == b.Mode()','(a.Mode() &^ 0111) == (b.Mode() &^ 0111)','./internal/envprofile','^TestMaterializeCaptureSameFragmentDrift/mode$'),
 ('M-provision-absent','internal/managerlock/managerlock.go','return manager.acquireHomeOnly(ctx, false)','if _, err := os.Stat(manager.Home()); os.IsNotExist(err) { return manager.AcquireHomeOnly(ctx, false) }; return manager.acquireHomeOnly(ctx, false)','./internal/envprofile','^TestMaterializeCaptureDoesNotProvision$'),
]
results=[]
for name,path,old,new,package,test in mutants:
 target=base/path;original=target.read_text()
 if original.count(old)!=1:raise RuntimeError((name,'mutation did not apply exactly once'))
 target.write_text(original.replace(old,new,1))
 log=root/f'.temp/SNAPLOCK-rev4/{name}.log'
 try:
  with log.open('w') as out:r=subprocess.run(['go','test',package,'-run',test,'-count=1','-timeout=2m'],cwd=base,stdout=out,stderr=subprocess.STDOUT,timeout=150)
  text=log.read_text()
  failures=[line.strip() for line in text.splitlines() if '--- FAIL:' in line]
  verdict='killed' if r.returncode==1 and failures and 'build failed' not in text else 'survived-or-unverified'
  results.append(dict(mutant=name,test=test,exit=r.returncode,verdict=verdict,failures=failures))
  print(json.dumps(results[-1]),flush=True)
 finally:target.write_text(original)
(root/'.temp/SNAPLOCK-rev4/mutants.json').write_text(json.dumps(results,indent=2)+'\n')

assert all(r["verdict"]=="killed" for r in results), results
