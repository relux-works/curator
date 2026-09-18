import sys,pathlib,copy,io,contextlib
sys.path.insert(0,'/tmp/2mglq0-review/regen/tools')
import validate
original=validate.load_json
path=validate.SUITE/'vectors/security-posture.json';base=original(path)
for mode in ['omit','misplace','nonclosed-output','nonclosed-input','consistent-rewrite']:
 v=copy.deepcopy(base);c=v['cases'][0]
 for key in ['curator_status_rows','env_status_rows']:
  rows=c['expected'][key];i=next(i for i,r in enumerate(rows) if r['gate']=='codex-seed')
  if mode=='omit':rows.pop(i)
  elif mode=='misplace':rows[i],rows[i+1]=rows[i+1],rows[i]
  elif mode in ['nonclosed-output','nonclosed-input']:rows[i]['value']='C'
  elif mode=='consistent-rewrite':rows[i]['value']='B'
 if mode=='nonclosed-input':c['shipped_revisions']['codex_seed']='C'
 if mode=='consistent-rewrite':c['shipped_revisions']['codex_seed']='B'
 validate.load_json=lambda p:v if p==path else original(p)
 err=io.StringIO()
 try:
  with contextlib.redirect_stderr(err),contextlib.redirect_stdout(io.StringIO()):code=validate.main()
 finally:validate.load_json=original
 print(mode, 'exit',code,err.getvalue().strip(),flush=True)
 assert code==1 and 'security-posture case' in err.getvalue()
print('main-entry refusal coverage 5/5; untouched on-disk digests; in-memory vector substitution at load_json only')
