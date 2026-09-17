import os, pathlib, subprocess, tempfile, json, shutil
root=pathlib.Path(tempfile.mkdtemp(prefix='TASK-260910-19w2aj-review-'))
env=os.environ.copy()
env.update(CURATOR_CONFIG=str(root/'home/config.json'), CURATOR_DRAFT_SOURCES_V1='1', GIT_CONFIG_GLOBAL='/dev/null', GIT_CONFIG_NOSYSTEM='1', GIT_AUTHOR_NAME='Review', GIT_AUTHOR_EMAIL='review@example.test', GIT_COMMITTER_NAME='Review', GIT_COMMITTER_EMAIL='review@example.test', XDG_CONFIG_HOME=str(root/'xdg'))
binary='/tmp/TASK-260910-19w2aj-review6-curator'
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
repo=root/'kit'; repo.mkdir()
for name in ['alpha','beta']:
    write(repo/'skills'/name/'SKILL.md','---\nname: '+name+'\ndescription: Test\n---\n# '+name+'\n')
    write(repo/'skills'/name/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{},'dependencies':{'skills':{}}}))
call(['git','-C',repo,'init','-q','-b','main']); call(['git','-C',repo,'add','.']); call(['git','-C',repo,'commit','-qm','two packages']); call(['git','-C',repo,'tag','v1'])
call(['git','-C',repo,'rm','-qr','skills/beta']); call(['git','-C',repo,'commit','-qm','remove beta at head'])
bare=root/'fixture.git'
call(['git','clone','--quiet','--bare',repo,bare])
real_git=shutil.which('git')
fake=root/'fake'; fake.mkdir()
wrapper='#!/bin/sh\nargs=""; for arg in "$@"; do\nif [ "$arg" = "https://fixture.test/kit.git" ]; then arg="file://'+str(bare)+'"; fi\nargs="$args\n$arg"; done\noldifs=$IFS; IFS="\n"; set -- $args; IFS=$oldifs\nexec "'+real_git+'" "$@"\n'
write(fake/'git',wrapper); (fake/'git').chmod(0o700)
env['PATH']=str(fake)+os.pathsep+env['PATH']
write(project/'Skillfile.json',json.dumps({'schema_version':2,'sources':{'s':{'git':'https://fixture.test/kit.git','tag':'v1'}},'skills':[{'from':'s','directory':'skills','include':['*']}]}))
print('CASE v1 collection contains alpha and beta; HEAD contains only alpha',flush=True)
cur('project','resolve','app')
locked=json.loads((project/'Skillfile.lock.json').read_text())
names=[m['name'] for m in locked['members']]
print('LOCKED MEMBERS', names, 'EXPECTED', ['alpha','beta'], flush=True)
assert names==['alpha']
print('FIXTURE',root,flush=True)
