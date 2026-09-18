import sys,copy,io,contextlib
from unittest.mock import patch
sys.path.insert(0,'tools')
import validate as v
original=v.load_json
base=original(v.SUITE/'vectors/security-posture.json')
def runmain(x):
 def load(p):
  return x if p == v.SUITE/'vectors/security-posture.json' else original(p)
 out=io.StringIO()
 with patch.object(v,'load_json',side_effect=load),contextlib.redirect_stdout(out),contextlib.redirect_stderr(out):
  rc=v.main()
 assert rc==1,out.getvalue()
 assert 'security-posture case' in out.getvalue(),out.getvalue()
 print(out.getvalue().strip(),flush=True)
x=copy.deepcopy(base)
c=next(c for c in x['cases'] if c['name']=='refusal-mcp-allowlist-empty-with-declarations')
c['machine']['security_posture']='permissive'
p,ps,e,s,n=v._posture_resolve(c['name'],c)
a=c['expected']; a.update(profile=p,profile_source=ps,effective=e,sources=s,passable_env_null_explicit=n,diagnostics=v._posture_diagnostics(c['name'],c,p,e,n),outcome='proceeds')
a['curator_status_rows'],a['env_status_rows']=v._posture_status_rows(c,p,ps,e,s)
runmain(x)
for target,donor in [('revision-A-default-permissive-status','locked-value-beats-explicit'),('revision-B-default-hardened-flip-install','schema1-machine-is-permissive')]:
 x=copy.deepcopy(base); d=copy.deepcopy(next(c for c in x['cases'] if c['name']==donor)); d['name']=target
 x['cases']=[d if c['name']==target else c for c in x['cases']];runmain(x)
refused=0
for t in base['cases']:
 for d in base['cases']:
  if t['name']==d['name']:continue
  x=copy.deepcopy(base);replacement=copy.deepcopy(d);replacement['name']=t['name']
  x['cases']=[replacement if c['name']==t['name'] else c for c in x['cases']]
  try:v.validate_security_posture_vectors(x)
  except v.ValidationFailure:refused+=1
  else:raise AssertionError((t['name'],d['name']))
print(f'Whole-case substitution sweep: {refused}/272 refused; main-entry replays: 3/3 refused')
