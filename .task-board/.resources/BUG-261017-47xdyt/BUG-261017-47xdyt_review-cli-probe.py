from pathlib import Path
import os,subprocess,json,hashlib,tempfile
root=Path(__file__).resolve().parent
box=Path(tempfile.mkdtemp(prefix='cli-',dir=root)); home=box/'home'; home.mkdir(mode=0o700)
project=box/'project'; project.mkdir()
binary=root/'curator'
env={'PATH':os.environ['PATH'],'HOME':str(home),'TMPDIR':str(box),'GIT_CONFIG_NOSYSTEM':'1','GIT_CONFIG_GLOBAL':os.devnull,'GIT_ALLOW_PROTOCOL':'file','XDG_CONFIG_HOME':str(home/'config'),'XDG_CACHE_HOME':str(home/'cache'),'XDG_DATA_HOME':str(home/'data')}
for k in ['CLAUDE_CONFIG_DIR','CODEX_HOME','OPENCODE_CONFIG_DIR','PI_CODING_AGENT_DIR']: env[k]=str(home/k)
def run(args,want=0):
 p=subprocess.run([str(x) for x in args],cwd=project,env=env,capture_output=True,text=True)
 text=p.stdout+p.stderr
 print(json.dumps({'args':[str(x).replace(str(box),'$FIXTURE').replace(str(root),'$REVIEW') for x in args],'exit':p.returncode,'output':text.replace(str(box),'$FIXTURE') }),flush=True)
 assert p.returncode==want,text
 return text
def put(p,v): p.parent.mkdir(parents=True,exist_ok=True);p.write_text(v)
def snapshot():
 out={}
 for base in [project/'.agents',project/'.claude',project/'.codex']:
  if not base.exists():continue
  for p in sorted(base.rglob('*')):
   rel=str(p.relative_to(project))
   if p.is_symlink():out[rel]=['link',os.readlink(p)]
   elif p.is_file():out[rel]=hashlib.sha256(p.read_bytes()).hexdigest()
   else:out[rel]='dir'
 return out
run(['git','init','-q','-b','main'])
put(project/'.gitignore','.agents/\n.claude/skills/\n.codex/skills/\n')
run([binary,'bootstrap','-non-interactive','--skills-root',box/'sources'])
manifest={'schema_version':2,'agents':['codex_cli','claude_code'],'sources':{'local':{'path':'./agents'}},'skills':[{'from':'local','directory':'skills','include':['*']}]}
put(project/'Skillfile.json',json.dumps(manifest))
for name in ['sample','stale']:
 put(project/'agents/skills'/name/'SKILL.md',f'---\nname: {name}\ndescription: fixture\n---\n# Fixture\n')
 put(project/'agents/skills'/name/'agent-skill.json','{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{}}}')
run([binary,'project','resolve','.']);run([binary,'install','.'])
# A: mix owned stale node and three unowned shapes.
manifest['skills'][0]['include']=['sample'];put(project/'Skillfile.json',json.dumps(manifest));run([binary,'project','refresh','.'])
store=project/'.agents/skills'
for rel in ['manual/SKILL.md','notes/note.md','loose.txt']:put(store/rel,'manual bytes\n')
link=project/'.codex/skills/foreign';link.symlink_to('../../.agents/skills/manual')
before=snapshot();out=run([binary,'install','.'],1)
for name in ['manual','notes','loose.txt']:assert str(store/name) in out
assert snapshot()==before;assert (link/'SKILL.md').read_text()=='manual bytes\n'
# Remove fixture conflicts and confirm real stale removal.
import shutil
for name in ['manual','notes']:shutil.rmtree(store/name)
(store/'loose.txt').unlink();link.unlink();run([binary,'install','.']);assert not (store/'stale').exists()
# C2 plus grammar free hunt, preserving all published files on every refusal.
live=store/'sample';saved=(live/'.csk-install.json').read_bytes()
for name,payload in [('missing',None),('duplicate-name',b'{"schema_version":5,"name":"other","name":"sample"}'),('string-schema',b'{"schema_version":"5","name":"sample"}'),('trailing-document',b'{"schema_version":5,"name":"sample"} {}')]:
 if payload is None:(live/'.csk-install.json').unlink()
 else:(live/'.csk-install.json').write_bytes(payload)
 put(live/'note.md','preserve me');before=snapshot();out=run([binary,'install','.'],1)
 assert str(live) in out;assert 'source_ownership_unproven' in out;assert snapshot()==before
 (live/'.csk-install.json').write_bytes(saved)
 print('HELD '+name,flush=True)
put(live/'SKILL.md','drift');run([binary,'install','.']);assert (live/'SKILL.md').read_text()!='drift'
run([binary,'install','.'])
print('PASS CLI A/C2, mixed stale cleanup, owned repair, clean reinstall, duplicate identity and trailing-document attacks')
