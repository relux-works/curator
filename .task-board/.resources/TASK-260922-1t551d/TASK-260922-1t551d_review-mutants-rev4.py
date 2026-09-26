from pathlib import Path
import subprocess
root=Path('.temp/TASK-260922-1t551d-review4')
p=root/'internal/envprofile/managed.go'
original=p.read_text()
cases=[
('skip-stale-removal','if err := removeStaleCredentialLinks(plan.adapter, plan.homeDir, native, prior, links); err != nil {','if err := removeStaleCredentialLinks(plan.adapter, plan.homeDir, native, nil, links); err != nil {','TestSharedToIsolatedRemovesStaleLink'),
('unlink-regular-file','return fmt.Errorf("%s: %s holds a regular file, expected a link to %s: refusing to remove or replace it", envregistry.DiagCredentialConflict, full, target)','_ = os.Remove(full); return os.Symlink(target, full)','TestCredentialLinkRegularFileRefuses'),
('silence-declared-dangling','v.warnings = append(v.warnings, fmt.Sprintf("passthrough entry %s is detached-pending: link target %s does not exist yet — log in to %s to populate it", path, target, plan.adapter.ID))','_ = target','TestReviewerDanglingExpectedTarget'),
('collapse-pending-into-conflict','v.warnings = append(v.warnings, fmt.Sprintf("passthrough entry %s is detached-pending: link target %s does not exist yet — log in to %s to populate it", path, target, plan.adapter.ID))','v.reasons = append(v.reasons, "environment_credential_conflict: missing target")','TestDanglingExpectedTargetRepairSucceeds'),
('auto-admitted','case "auto":\n\t\treturn fmt.Errorf("%s: isolated is unsupported','case "auto":\n        return nil\n\t\treturn fmt.Errorf("%s: isolated is unsupported','TestReviewerLiteralCodexSelector'),
('unknown-admitted','switch store {\n\tcase "file", "keyring", "auto":','if store == "ephemeral" { return "file", nil }; switch store {\n\tcase "file", "keyring", "auto":','TestReviewerLiteralCodexSelector'),
('malformed-as-absent','return "", fmt.Errorf("codex native config.toml is not valid TOML: %v", err)','return "file", nil','TestCodexCredentialStoreTOMLSpellings'),
('foreign-relink','if !recorded {','if !recorded && got == "" {','TestCredentialLinkUnrecordedSymlinkRefuses'),
]
with Path('.temp/TASK-260922-1t551d-mutants.log').open('w') as log:
 for name,old,new,test in cases:
  assert original.count(old)==1,(name,original.count(old))
  try:
   p.write_text(original.replace(old,new))
   r=subprocess.run(['go','test','./internal/envprofile','-run','^'+test+'$','-count=1','-timeout=90s'],cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=110)
   log.write(f'## {name}\nexit={r.returncode}\n{r.stdout}\n');log.flush()
   print(name,r.returncode,flush=True)
  finally: p.write_text(original)
assert p.read_text()==original
