import json,pathlib,subprocess,hashlib
root=pathlib.Path(pathlib.Path('/tmp/s4-landing-root').read_text())/'candidate'
python='/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python3'
s4='conformance/v1/vectors/environments-env-passthrough.json'
s6='conformance/v1/vectors/shell-hook-trust.json'
e4='conformance/v1/vectors/umbrella-provider-resolution.json'
def case(v,group,name):return next(c for c in v[group] if c['name']==name)
def unlisted(v):
 c=case(v,'default_resolution_cases','s4-enforce-absent-drops-all');c['passed']=['FIGMA_API_KEY'];c['dropped']=[]
def invalid(v):
 c=v['surfacing_cases'][0];c['expected_bytes']='INVALID\n';c['expected_byte_length']=8;c['expected_sha256']='sha256:'+hashlib.sha256(b'INVALID\n').hexdigest()
def sourced(v):case(v,'cases','changed-env-sh-B-enforcing-not-sourced')['sourced']=True
def resolution(v):case(v,'cases','s6-planted-path-provider-warns-then-refuses')['revision_b']={'resolved':'/home/operator/work/acme/.bin/curator-run','diagnostic':None}
def refresh_pins():
 manifest=root/'conformance/v1/manifest.json';m=json.loads(manifest.read_bytes())
 for item in m['files']:item['sha256']='sha256:'+hashlib.sha256((root/'conformance/v1'/item['path']).read_bytes()).hexdigest()
 manifest.write_text(json.dumps(m,indent=2,sort_keys=True)+'\n')
 digest='sha256:'+hashlib.sha256(manifest.read_bytes()).hexdigest()
 release=root/'release/1.0.0-rc.9.json';v=json.loads(release.read_bytes())
 v['candidate_protocol_pin']['manifest_sha256']=digest;v['downstream_consumption']['required_manifest_sha256']=digest
 release.write_text(json.dumps(v,indent=2,sort_keys=True)+'\n')
for name,rel,mutate in [('unlisted-passed',s4,unlisted),('invalid-output',s4,invalid),('sourced',s6,sourced),('resolution',e4,resolution)]:
 p=root/rel;original=p.read_bytes()
 try:
  v=json.loads(original);mutate(v);p.write_text(json.dumps(v,indent=2)+'\n')
  refresh_pins()
  assert json.loads(p.read_bytes())==v
  result=subprocess.run([python,'-B','tools/validate.py'],cwd=root,capture_output=True,text=True)
  print(name,'exit='+str(result.returncode),result.stdout,result.stderr,flush=True)
  assert result.returncode==1 and 'validation failed:' in result.stderr
 finally:p.write_bytes(original)
subprocess.run(['go','run','./tools/generate-vectors','-root','.'],cwd=root,check=True,capture_output=True)
subprocess.run(['git','diff','--exit-code'],cwd=root,check=True)
print('mutants rejected: 4/4; restored scratch diff exit=0',flush=True)
