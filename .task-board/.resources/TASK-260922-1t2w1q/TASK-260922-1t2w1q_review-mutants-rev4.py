from pathlib import Path
import subprocess
root=Path(__file__).resolve().parent/'candidate'
p=root/'internal/envprofile/migrate.go'
original=p.read_text()
mutants=[('validate-only-done-links', 'TestMigrateRecoveryRefusesUnexpectedTarget', original.replace('for _, op := range journal.Ops {\n\t\tfull :=', 'for _, op := range journal.Ops {\n\t\tif !op.Done { continue }\n\t\tfull :=',1)), ('glob-cleanup', 'TestReviewerRecoveryPreservesRegularTemp', original.replace('\t\top := journal.Ops[i]\n', '\t\top := journal.Ops[i]\n\t\tmatches, _ := filepath.Glob(filepath.Join(ManagedHomeDir(home, op.Profile, op.EnvID), ".migrate-*.tmp"))\n\t\tfor _, match := range matches { _ = os.Remove(match) }\n',1))]
try:
 for name,test,mutated in mutants:
  assert mutated != original
  p.write_text(mutated)
  result=subprocess.run(['go','test','./internal/envprofile','-run','^'+test+'$','-count=1','-timeout=90s','-v'],cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  print(name, 'EXIT',result.returncode,flush=True); print(result.stdout,flush=True)
  assert result.returncode!=0 and '--- FAIL: '+test in result.stdout
  p.write_text(original)
finally: p.write_text(original)
