from pathlib import Path
import subprocess
root=Path('.temp/review-1t551d'); p=root/'internal/envprofile/managed.go'; original=p.read_text()
cases=[
('skip-recorded-removal','if prior == nil || prior.Passthrough == nil {','if true || prior == nil || prior.Passthrough == nil {','TestSharedToIsolatedRemovesStaleLink'),
('unconditional-remove','if err := ensureCredentialLink(plan.homeDir, path, links[path]); err != nil {','_ = os.Remove(filepath.Join(plan.homeDir, filepath.FromSlash(path)))\n\t\tif err := ensureCredentialLink(plan.homeDir, path, links[path]); err != nil {','TestCredentialLinkRegularFileRefuses'),
('narrow-target-basename','if got != target {','if filepath.Base(got) != filepath.Base(target) {','TestDanglingPiLinkReportedDetached'),
]
for name,old,new,test in cases:
 assert old in original
 try:
  p.write_text(original.replace(old,new,1))
  r=subprocess.run(['go','test','./internal/envprofile','-run','^'+test+'$','-count=1'],cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=120)
  text=f'{name}: exit={r.returncode}\n{r.stdout}'
  print(text,flush=True)
  (root.parent/('review-1t551d-'+name+'.log')).write_text(text)
 finally: p.write_text(original)
