import os, pathlib, subprocess, tempfile, json, shutil
root=pathlib.Path(tempfile.mkdtemp(prefix='TASK-260910-19w2aj-review-'))
env=os.environ.copy()
env.update(CURATOR_CONFIG=str(root/'home/config.json'), CURATOR_DRAFT_SOURCES_V1='1', GIT_CONFIG_GLOBAL='/dev/null', GIT_CONFIG_NOSYSTEM='1', GIT_AUTHOR_NAME='Review', GIT_AUTHOR_EMAIL='review@example.test', GIT_COMMITTER_NAME='Review', GIT_COMMITTER_EMAIL='review@example.test', XDG_CONFIG_HOME=str(root/'xdg'))
binary='/tmp/TASK-260910-19w2aj-review-cli'
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
cur('bootstrap','--non-interactive','--skills-root',root/'skills-root'); cur('project','add','app',project,'--agents','codex_cli'); call(['git','-C',project,'init','-q','-b','main'])
repo=root/'skills-root/review'
write(repo/'SKILL.md','---\nname: review\ndescription: Test\n---\n# review\n')
write(repo/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{},'dependencies':{'skills':{}}}))
call(['git','-C',repo,'init','-q','-b','main']); call(['git','-C',repo,'add','.']); call(['git','-C',repo,'commit','-qm','fixture']); call(['git','-C',repo,'tag','v1'])
for kind in ['configured','network']:
    if kind=='network': call(['git','-C',repo,'remote','add','origin','https://fixture.test/review.git'])
    skill={'name':'review','tag':'v1'}
    if kind=='network': skill['git']='https://fixture.test/review.git'
    write(project/'Skillfile.json',json.dumps({'schema_version':2,'skills':[skill]}))
    print('CASE',kind,flush=True)
    cur('project','resolve','app')
    cur('install','app','--dry-run',expected=None)
    cur('install','app',expected=None)
print('FIXTURE',root,flush=True)
