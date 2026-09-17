import os, pathlib, subprocess, json, tempfile, sys
base=pathlib.Path(__file__).resolve().parent
binary=base/'curator'
root=pathlib.Path(tempfile.mkdtemp(prefix='fixture-',dir=base))
env=dict(os.environ,CURATOR_CONFIG=str(root/'home/config.json'),CURATOR_DRAFT_SOURCES_V1='1',GIT_CONFIG_NOSYSTEM='1',GIT_CONFIG_GLOBAL='/dev/null',GIT_TERMINAL_PROMPT='0')
def run(*args,cwd=None):
 p=subprocess.run(list(map(str,args)),cwd=cwd,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=30)
 print('COMMAND',*args,'EXIT',p.returncode,'\n'+p.stdout,flush=True)
 assert p.returncode==0
 return p.stdout.strip()
def git(repo,*args): return run('git','-C',repo,*args)
project=root/'project';project.mkdir()
upstream=root/'upstream';upstream.mkdir()
git(upstream,'init','-q','-b','main');git(upstream,'config','user.name','Review');git(upstream,'config','user.email','review@example.test')
(upstream/'SKILL.md').write_text('---\nname: review\ndescription: Test\n---\n# review\n')
(upstream/'agent-skill.json').write_text('{"schema_version":4,"capabilities":{},"commands":{"tool":{"type":"script","unix_path":"scripts/tool.sh"}},"dependencies":{"skills":{}},"runtime_roots":["scripts"]}')
(upstream/'scripts').mkdir(); script=upstream/'scripts/tool.sh';script.write_text('#!/bin/sh\necho original\n');script.chmod(0o755)
git(upstream,'add','.');git(upstream,'commit','-qm','initial');old=git(upstream,'rev-parse','HEAD')
run(binary,'bootstrap','--non-interactive','--skills-root',root/'skills-root')
run(binary,'project','add','app',project,'--agents','codex_cli');git(project,'init','-q','-b','main')
checkout=root/'skills-root/review';run('git','clone',upstream,checkout)
(project/'Skillfile.json').write_text(json.dumps({'schema_version':2,'skills':[dict(name='review',branch='main',**({'git':'https://fixture.test/review.git'} if '--network' in sys.argv else {}))]}))
run(binary,'project','resolve','app')
def locked(): return json.loads((project/'Skillfile.lock.json').read_text())['members'][0]['package']['commit']['hex']
assert locked()==old
run(binary,'install','app')
script.write_text('#!/bin/sh\necho refreshed\n');git(upstream,'add','.');git(upstream,'commit','-qm','runtime-only change');new=git(upstream,'rev-parse','HEAD')
run(binary,'project','refresh','app')
print('OLD',old,'REMOTE_NEW',new,'REFRESH_LOCK',locked(),flush=True)
assert locked()==old and old!=new, 'expected reproduction: stale pin'
print('REPRODUCED: refresh exit 0 but failed to acquire runtime-only branch change',flush=True)
git(checkout,'fetch','origin');run(binary,'project','refresh','app');assert locked()==new
print('CONTROL: after manual fetch refresh locks the new commit',flush=True)
