from pathlib import Path
import subprocess
root=Path('.temp/review2-code').resolve()
def run(name,file,old,new,mask,expect):
 p=root/file
 original=p.read_text()
 assert original.count(old)==1,(name,original.count(old))
 try:
  p.write_text(original.replace(old,new))
  r=subprocess.run(['go','test','./internal/envprofile','-run',mask,'-count=1','-timeout=90s','-v'],cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  out=f'{name}: exit={r.returncode}, expected={expect}\n'+r.stdout
  print(out,flush=True)
  Path('.temp/'+name+'.log').write_text(out)
 finally:
  p.write_text(original)
run('review-state-secret-copy','internal/envprofile/migrate.go',
 '\t\tcase MigrateOpRelink:\n\t\t\tif err := os.MkdirAll(filepath.Dir(full), 0o755);',
 '\t\tcase MigrateOpRelink:\n\t\t\tif payload, err := os.ReadFile(op.To); err == nil {\n\t\t\t\tif err := os.WriteFile(filepath.Join(filepath.Dir(migrationJournalPath(req.Home)), "credential-backup"), payload, 0o600); err != nil { return applied, err }\n\t\t\t}\n\t\t\tif err := os.MkdirAll(filepath.Dir(full), 0o755);',
 '^TestMigrateNoSecretCopies$',1)

run('review-interrupted-state-copy','internal/envprofile/migrate.go',
 '\t\t\tif err := req.InjectFault(MigrateFaultCrash, len(applied)); err != nil {',
 '\t\t\tif payload, err := os.ReadFile(op.From); err == nil {\n\t\t\t\tif err := os.WriteFile(filepath.Join(filepath.Dir(migrationJournalPath(req.Home)), "transient-backup"), payload, 0o600); err != nil { return applied, err }\n\t\t\t}\n\t\t\tif err := req.InjectFault(MigrateFaultCrash, len(applied)); err != nil {',
 '^TestMigrateNoSecretCopies$',1)
