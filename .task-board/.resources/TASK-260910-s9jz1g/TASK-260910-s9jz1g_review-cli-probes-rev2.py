import os, subprocess, tempfile, json
from pathlib import Path
from csk_registry.signing import load_key
cli='/tmp/s9-review2/venv/bin/curator-skill-registry'
var='CSK_REGISTRY_KEY_PASSPHRASE'
env=dict(os.environ); env.pop(var,None)
def run(home,args,secret=None,fail=False):
 e=dict(env)
 if secret is not None: e[var]=secret
 p=subprocess.run([cli,'--home',str(home),*args],env=e,text=True,capture_output=True,timeout=15)
 if fail:
  assert p.returncode==1,(args,p.stdout,p.stderr)
  assert p.stderr.count(var)==1 and len(p.stderr.splitlines())==1
  assert 'Traceback' not in p.stderr and 'PRIVATE KEY' not in p.stderr
  assert 'review-secret-unique' not in p.stderr
 else: assert p.returncode==0,(args,p.stderr)
 print(args[0], 'refused' if fail else 'passed',p.returncode)
 return p
for initial in (None,'review-secret-unique'):
 h=Path(tempfile.mkdtemp(prefix='s9-cli-'))
 run(h,['genkey'],initial); run(h,['prepare-key-rotation'],initial)
 staged=load_key((h/'next-signing-key.pem').read_bytes(),initial.encode() if initial else None)
 run(h,['activate-key-rotation','--confirm-pins-deployed'],'review-secret-unique')
 pem=(h/'signing-key.pem').read_bytes()
 assert pem.startswith(b'-----BEGIN ENCRYPTED PRIVATE KEY')
 assert load_key(pem,b'review-secret-unique').public_pinned==staged.public_pinned
 assert not (h/'next-signing-key.pem').exists()
 assert (h/'signing-key.pem').stat().st_mode & 0o777 == 0o600
 for secret in ('',None,'different-secret'):
  before={p.name:p.read_bytes() for p in h.iterdir() if p.is_file()}
  for cmd in (['prepare-key-rotation'],['serve']): run(h,cmd,secret,True)
  assert before=={p.name:p.read_bytes() for p in h.iterdir() if p.is_file()}
h=Path(tempfile.mkdtemp(prefix='s9-empty-'))
run(h,['genkey'],'',True); assert not list(h.iterdir())
print('CLI probes PASS: both activation shapes; 12 encrypted-load refusals; empty genkey; no key bytes or secret in diagnostics')
h=Path(tempfile.mkdtemp(prefix='s9-empty-rotation-'))
run(h,['genkey']); run(h,['prepare-key-rotation'])
before={p.name:p.read_bytes() for p in h.iterdir() if p.is_file()}
for cmd in (['activate-key-rotation','--confirm-pins-deployed'],['cancel-key-rotation','--confirm'],['prepare-key-rotation'],['serve']):
 run(h,cmd,'',True)
 assert before=={p.name:p.read_bytes() for p in h.iterdir() if p.is_file()}
print('Empty configured secret: staged rotation activate/cancel/prepare and serve refuse without file changes: 4/4')
