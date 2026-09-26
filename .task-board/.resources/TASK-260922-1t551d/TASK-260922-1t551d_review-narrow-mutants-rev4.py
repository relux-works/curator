from pathlib import Path
import subprocess
root=Path('.temp/TASK-260922-1t551d-review4'); p=root/'internal/envprofile/managed.go'; original=p.read_text()
cases=[('empty-file-exempt', 'return fmt.Errorf("%s: %s holds a regular file, expected a link to %s: refusing to remove or replace it", envregistry.DiagCredentialConflict, full, target)', 'if info.Size() == 0 { _ = os.Remove(full); return os.Symlink(target, full) }; return fmt.Errorf("%s: %s holds a regular file, expected a link to %s: refusing to remove or replace it", envregistry.DiagCredentialConflict, full, target)', 'TestCredentialLinkRegularFileRefuses'),('wrong-target-basename-only','if got != target {','if filepath.Base(got) != filepath.Base(target) {','TestDanglingPiLinkReportedDetached'),('stale-target-basename-only','if got != declared {','if filepath.Base(got) != filepath.Base(declared) {','TestStaleCredentialLinkRefusals')]
with Path('.temp/TASK-260922-1t551d-narrow-mutants.log').open('w') as log:
 for name,old,new,test in cases:
  assert original.count(old)==1
  try:
   p.write_text(original.replace(old,new))
   r=subprocess.run(['go','test','./internal/envprofile','-run','^'+test+'$','-count=1','-timeout=90s'],cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=110)
   log.write(f'## {name}\nexit={r.returncode}\n{r.stdout}\n'); log.flush();print(name,r.returncode,flush=True)
  finally:p.write_text(original)
