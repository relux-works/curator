import os, pathlib, subprocess, tempfile, json, shutil
root=pathlib.Path(tempfile.mkdtemp(prefix='TASK-260910-19w2aj-review-'))
env=os.environ.copy()
env.update(CURATOR_CONFIG=str(root/'home/config.json'), CURATOR_DRAFT_SOURCES_V1='1', GIT_CONFIG_GLOBAL='/dev/null', GIT_CONFIG_NOSYSTEM='1', GIT_AUTHOR_NAME='Review', GIT_AUTHOR_EMAIL='review@example.test', GIT_COMMITTER_NAME='Review', GIT_COMMITTER_EMAIL='review@example.test', XDG_CONFIG_HOME=str(root/'xdg'))
binary='/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/.temp/review-19w2aj-rev4/curator'
def call(args, expected=0):
    p=subprocess.run([str(x) for x in args],env=env,capture_output=True,text=True)
    print('COMMAND', ' '.join(str(x) for x in args), 'EXIT', p.returncode, flush=True)
    print(p.stdout+p.stderr,flush=True)
    if expected is not None: assert p.returncode==expected
    return p.returncode
def cur(*args,**kwargs): return call([binary,*args],**kwargs)
def write(path,text):
    path.parent.mkdir(parents=True,exist_ok=True); path.write_text(text)
project=root/'project'; project.mkdir()
repo=root/'repo'; pkg=repo/'skills/review'
write(pkg/'SKILL.md','---\nname: review\ndescription: Test\n---\n# review\n')
write(pkg/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{'tool':{'type':'script','unix_path':'scripts/tool.sh'}},'dependencies':{'skills':{}},'runtime_roots':['scripts']}))
write(pkg/'scripts/tool.sh','#!/bin/sh\necho original\n'); (pkg/'scripts/tool.sh').chmod(0o755)
call(['git','-C',repo,'init','-q','-b','main']); call(['git','-C',repo,'add','.']); call(['git','-C',repo,'commit','-qm','fixture']); call(['git','-C',repo,'tag','v1']); call(['git','-C',repo,'tag','v2'])
bare=root/'kit.git'; call(['git','clone','--bare','--quiet',repo,bare])
commit=subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD'],env=env,text=True).strip()
cur('bootstrap','--non-interactive','--skills-root',root/'skills-root'); cur('project','add','app',project,'--agents','codex_cli'); call(['git','-C',project,'init','-q','-b','main'])
write(project/'Skillfile.json',json.dumps({'schema_version':2,'sources':{'a':{'git':'https://fixture.test/kit.git','tag':'v1'},'s':{'git':'https://fixture.test/kit.git','tag':'v2'}},'skills':[{'name':'review','from':'s','directory':'skills/review'}]}))
realgit=shutil.which('git'); fake=root/'fake/git'
write(fake,'#!/usr/bin/python3\nimport os,sys\na=["file://'+str(bare)+'" if x=="https://fixture.test/kit.git" else x for x in sys.argv[1:]]\nos.execv("'+realgit+'",["git"]+a)\n'); fake.chmod(0o755)
env['PATH']=str(fake.parent)+os.pathsep+env['PATH']
cur('project','resolve','app')
cur('install','app','--dry-run')
print('BASELINE REAL INSTALL', flush=True)
cur('install','app',expected=None)

for p in (project/".agents/skills/review").rglob("*"):
    if p.is_file() and p.name == ".csk-install.json": print("MARKER",p,p.read_text())
print("FIXTURE",root)

selected=json.loads((project/"Skillfile.json").read_text()); recorded=json.loads((project/".agents/skills/review/.csk-install.json").read_text()); expected=selected["sources"][selected["skills"][0]["from"]]["tag"]; print("SELECTED TAG",expected,"RECORDED TAG",recorded["ref"]); assert expected=="v2" and recorded["ref"]=="v1"
