from pathlib import Path
import subprocess,json,hashlib,re
r=Path(Path('/tmp/s4-landing-root').read_text());src=Path('/Users/administrator/Developer/ReluxWorks/.worktrees/curator-spec-s4-passthrough')
def blob(rev,f):return subprocess.check_output(['git','show',rev+':'+f],cwd=src)
files=subprocess.check_output(['git','ls-files'],cwd=src,text=True).splitlines()
assert all(hashlib.sha1(b'blob '+str(len((src/f).read_bytes())).encode()+b'\0'+(src/f).read_bytes()).hexdigest()==oid for f,oid in [(row.split('\t',1)[1],row.split()[2]) for row in subprocess.check_output(['git','ls-tree','-r','23dafa7'],cwd=src,text=True).splitlines()] )
print('Exact head byte coverage:',len(files),'/',len(files),'tracked files; delivery worktree equals commit')
f='conformance/v1/vectors/manager-config-v2.json'
a=json.loads((r/'accepted'/f).read_bytes());b=json.loads(blob('0da4020',f));c=json.loads((src/f).read_bytes())
def byname(v):return {x['name']:x for x in v}
aa,bb,cc=map(byname,(a,b,c))
assert set(cc)==set(aa)|set(bb)
for name,old in bb.items():
 expected=json.loads(json.dumps(old))
 env=expected.get('expected',{}).get('environments')
 if env is not None and 'passable_env_names' not in expected['input'].get('environments',{}):env['passable_env_names']=[]
 assert cc[name]==expected,name
for name in set(aa)-set(bb):
 expected=json.loads(json.dumps(aa[name]));env=expected.get('expected',{}).get('environments')
 if env is not None:env.update(transitive_system_modules='drop',system_module_waivers=[],provider_directories=[])
 assert cc[name]==expected,name
print('Manager case inventory union exact:',len(aa),'+',len(bb),'=>',len(cc),'; all landed cases preserved except required absent default; all new S4 cases inherit E2/E4 knobs')
for f in ['cli/curator.md','conformance/v1/vectors/shell-hook-trust.json','conformance/v1/vectors/umbrella-provider-resolution.json','schemas/v1/system-config-v2.schema.json']:
 assert (src/f).read_bytes()==blob('0da4020',f)
 print('Landed byte-identical:',f)
# Unchanged headings/sections preserve each family's authoritative rule once.
env=(src/'protocol/environments.md').read_text()
for section in ['## 3. Context modules','### 5.5 System-prompt output','## 11. Umbrella subcommand discovery','### 12.2 Lockable knobs']:
 def get(s):
  part=s[s.index(section):];m=re.search(r'\n#{1,'+str(len(section.split()[0]))+r'} ',part)
  return part[:m.start()] if m else part
 assert get(env)==get(blob('0da4020','protocol/environments.md').decode()),section
 assert env.count(section+'\n')==1
 print('Landed section byte-identical and single:',section)
print('Fidelity assertions passed')
