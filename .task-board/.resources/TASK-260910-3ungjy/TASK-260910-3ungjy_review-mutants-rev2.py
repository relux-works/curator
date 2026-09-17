import pathlib,subprocess,os
root=pathlib.Path('/tmp/TASK-260910-3ungjy-attacks')
f=root/'cmd/curator/hook.go'; original=f.read_bytes()
env=dict(os.environ,CURATOR_CONFORMANCE_ROOT='/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1')
mutants=[('M1 reapproval only reuses old digest for operator records', 'existing.SHA256 == hookapproval.Digest(payload) && existing.ApprovedBy', 'existing.ApprovedBy', 'TestHookApproveReRecordsAfterChange'),('M2 changed-check only covers manager approvals','if row.State == hookapproval.PostureChanged {','if row.State == hookapproval.PostureChanged && row.ApprovedBy == hookapproval.ApprovedByManager {','TestStatusReportsShellHookTrustPosture')]
try:
 for name,before,after,test in mutants:
  text=original.decode(); assert text.count(before)==1
  f.write_text(text.replace(before,after))
  print(name,flush=True)
  r=subprocess.run(['go','test','-count=1','-timeout','3m','./cmd/curator','-run','^'+test+'$'],cwd=root,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=210)
  print(r.stdout, 'exit='+str(r.returncode),flush=True)
  assert r.returncode==1,'mutant survived or infrastructure error'
  f.write_bytes(original); assert f.read_bytes()==original
finally: f.write_bytes(original)
r=subprocess.run(['go','test','-count=1','-timeout','3m','./cmd/curator','-run','^(TestHookApproveReRecordsAfterChange|TestStatusReportsShellHookTrustPosture)$'],cwd=root,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=210)
print('restored:',r.stdout,'exit='+str(r.returncode),flush=True)
assert r.returncode==0
print('2/2 narrowing mutants killed; byte-identical restore',flush=True)
