import os, pathlib, subprocess, tempfile, json, shutil
root=pathlib.Path(tempfile.mkdtemp(prefix='curator-review7-allowlist-'))
env=os.environ.copy()
env.update(CURATOR_CONFIG=str(root/'home/config.json'), CURATOR_DRAFT_SOURCES_V1='1', GIT_CONFIG_GLOBAL='/dev/null', GIT_CONFIG_NOSYSTEM='1', GIT_AUTHOR_NAME='Review', GIT_AUTHOR_EMAIL='review@example.test', GIT_COMMITTER_NAME='Review', GIT_COMMITTER_EMAIL='review@example.test', XDG_CONFIG_HOME=str(root/'xdg'))
binary='/tmp/curator-review-19w2aj-rev7'
def call(args, expected=0):
 p=subprocess.run([str(x) for x in args],env=env,cwd=root,capture_output=True,text=True,timeout=60)
 print('COMMAND', ' '.join(str(x) for x in args), 'EXIT',p.returncode,flush=True); print(p.stdout+p.stderr,flush=True)
 if expected is not None: assert p.returncode==expected
 return p
def cur(*args,**kw): return call([binary,*args],**kw)
def write(path,text):
 path.parent.mkdir(parents=True,exist_ok=True); path.write_text(text)
def skill(path,name,deps):
 write(path/'SKILL.md','---\nname: '+name+'\ndescription: Test\n---\n# '+name+'\n')
 write(path/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{},'dependencies':{'skills':deps}}))
repo=root/'provider-repo'; skill(repo,'provider',{})
call(['git','-C',repo,'init','-q','-b','main']); call(['git','-C',repo,'add','.']); call(['git','-C',repo,'commit','-qm','fixture']); call(['git','-C',repo,'tag','v1'])
project=root/'project'; project.mkdir()
cur('bootstrap','--non-interactive','--skills-root',root/'skills-root'); cur('project','add','app',project,'--agents','codex_cli'); call(['git','-C',project,'init','-q','-b','main'])
cfgpath=root/'home/config.json'; cfg=json.loads(cfgpath.read_text()); cfg['allowed_sources']=['trusted.test/team']; write(cfgpath,json.dumps(cfg))
realgit=shutil.which('git'); fake=root/'fake/git'; traffic=root/'acquisition.log'
write(fake,'#!/usr/bin/python3\nimport os,sys\na=sys.argv[1:]\nif "clone" in a: open('+repr(str(traffic))+',"a").write(repr(a)+"\\n")\na=["file://'+str(repo)+'" if x=="https://denied.test/provider.git" else x for x in a]\nos.execv('+repr(realgit)+',["git"]+a)\n'); fake.chmod(0o755); env['PATH']=str(fake.parent)+os.pathsep+env['PATH']
# Legacy control refuses before any clone.
write(project/'Skillfile.json',json.dumps({'schema_version':1,'skills':[{'name':'provider','git':'https://denied.test/provider.git','tag':'v1'}]}))
p=cur('install','app','--dry-run',expected=1); assert 'source not allowed' in p.stderr+p.stdout; assert not traffic.exists()
# Draft alias control also refuses before clone.
write(project/'Skillfile.json',json.dumps({'schema_version':2,'sources':{'s':{'git':'https://denied.test/provider.git','tag':'v1'}},'skills':[{'name':'provider','from':'s','directory':'.'}]}))
p=cur('project','resolve','app',expected=1); assert 'source not allowed' in p.stderr+p.stdout; assert not traffic.exists()
# Draft unchanged legacy root bypasses the same configured allowlist.
write(project/'Skillfile.json',json.dumps({'schema_version':2,'skills':[{'name':'provider','git':'https://denied.test/provider.git','tag':'v1'}]}))
cur('project','resolve','app'); assert traffic.exists(); print('DENIED LEGACY CLONE',traffic.read_text(),flush=True)
shutil.rmtree(root/'skills-root/provider'); traffic.unlink()
# Draft local root's transitive dependency bypasses it as well.
consumer=root/'local/review'; skill(consumer,'review',{'provider':{'git':'https://denied.test/provider.git','ref':{'kind':'tag','value':'v1'}}})
write(project/'Skillfile.json',json.dumps({'schema_version':2,'sources':{'s':{'path':str(root/'local')}},'skills':[{'name':'review','from':'s','directory':'review'}]}))
cur('project','resolve','app'); assert traffic.exists(); print('DENIED TRANSITIVE CLONE',traffic.read_text(),flush=True)
lock=json.loads((project/'Skillfile.lock.json').read_text()); assert any(m['name']=='provider' for m in lock['members']); print('LOCK MEMBERS',json.dumps(lock['members']),flush=True)
print('FIXTURE',root,flush=True)
# Positive control: explicitly allow the provider and resolve from a clean checkout location.
shutil.rmtree(root/'skills-root/provider'); traffic.unlink()
cfg['allowed_sources']=['denied.test']; write(cfgpath,json.dumps(cfg))
cur('project','resolve','app'); assert traffic.exists()
print('POSITIVE CONTROL: allowed transitive acquisition succeeds',flush=True)
