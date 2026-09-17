import os, pathlib, subprocess, tempfile, json, shutil
root=pathlib.Path(tempfile.mkdtemp(prefix='TASK-260910-19w2aj-review-'))
env=os.environ.copy()
env.update(CURATOR_CONFIG=str(root/'home/config.json'), CURATOR_DRAFT_SOURCES_V1='1', GIT_CONFIG_GLOBAL='/dev/null', GIT_CONFIG_NOSYSTEM='1', GIT_AUTHOR_NAME='Review', GIT_AUTHOR_EMAIL='review@example.test', GIT_COMMITTER_NAME='Review', GIT_COMMITTER_EMAIL='review@example.test', XDG_CONFIG_HOME=str(root/'xdg'))
binary='/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/.temp/review-19w2aj-rev4/curator'
def call(args, expected=0):
    p=subprocess.run([str(x) for x in args],env=env,cwd=root,capture_output=True,text=True)
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
call(['git','-C',repo,'init','-q','-b','main']); call(['git','-C',repo,'add','.']); call(['git','-C',repo,'commit','-qm','fixture']); call(['git','-C',repo,'tag','v1'])
bare=root/'kit.git'; call(['git','clone','--bare','--quiet',repo,bare])
commit=subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD'],env=env,text=True).strip()
cur('bootstrap','--non-interactive','--skills-root',root/'skills-root'); cur('project','add','app',project,'--agents','codex_cli'); call(['git','-C',project,'init','-q','-b','main'])
write(project/'Skillfile.json',json.dumps({'schema_version':2,'sources':{'s':{'git':'https://fixture.test/kit.git','tag':'v1'}},'skills':[{'name':'review','from':'s','directory':'skills/review'}]}))
realgit=shutil.which('git'); fake=root/'fake/git'
write(fake,'#!/usr/bin/python3\nimport os,sys\na=["file://'+str(bare)+'" if x=="https://fixture.test/kit.git" else "file:///nonexistent-review-provider" if x=="https://fixture.test/provider.git" else x for x in sys.argv[1:]]\nos.execv("'+realgit+'",["git"]+a)\n'); fake.chmod(0o755)
env['PATH']=str(fake.parent)+os.pathsep+env['PATH']

provider=root/'skills-root/provider'
shutil.copytree(pkg,provider)
write(provider/'SKILL.md','---\nname: provider\ndescription: Test\n---\n# provider\n')
write(provider/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{},'dependencies':{'skills':{}}}))
call(['git','-C',provider,'init','-q','-b','main']); call(['git','-C',provider,'add','.']); call(['git','-C',provider,'commit','-qm','provider']); call(['git','-C',provider,'tag','v1'])
consumer=root/'local/review'; shutil.copytree(pkg,consumer)
write(consumer/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{},'dependencies':{'skills':{'provider':{'git':'https://fixture.test/provider.git','ref':{'kind':'tag','value':'v1'}}}}}))
write(project/'Skillfile.json',json.dumps({'schema_version':2,'sources':{'s':{'path':str(root/'local')}},'skills':[{'name':'review','from':'s','directory':'review'}]}))
print('CONFIGURED PROVIDER',provider,flush=True)
cur('project','resolve','app',expected=1)
(root/'provider').symlink_to(provider, target_is_directory=True)
print('CONTROL: expose the configured provider under process CWD',flush=True)
cur('project','resolve','app',expected=0)
print('FIXTURE',root,flush=True)
