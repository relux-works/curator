import os, json, pathlib, subprocess, hashlib, shlex

BASE = pathlib.Path(__file__).resolve().parent
ROOT = BASE / 'sandbox-05'
ROOT.mkdir()
BIN = BASE / 'curator'
LOG = ROOT / 'sandbox.log'
LOG.write_text('Private CLI reproductions; no inherited environment or credentials.\n')
def log(s):
    with LOG.open('a') as f: f.write(str(s)+'\n')
def write(p, s):
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(s)
def new(name):
    root = ROOT / name
    root.mkdir()
    home = root / 'home'
    home.mkdir(mode=0o700)
    project = root / 'project'
    project.mkdir()
    write(project/'.gitignore','.agents/\n.claude/skills/\n.codex/skills/\n')
    env = {'HOME':str(home), 'PATH':'/usr/bin:/bin', 'TMPDIR':str(root/'tmp'),
           'CLAUDE_CONFIG_DIR':str(home/'.claude'), 'CODEX_HOME':str(home/'.codex'),
           'XDG_CONFIG_HOME':str(home/'.config'), 'XDG_DATA_HOME':str(home/'.local/share'),
           'PI_CODING_AGENT_DIR':str(home/'.pi'), 'GIT_CONFIG_NOSYSTEM':'1',
           'GIT_CONFIG_GLOBAL':'/dev/null', 'GIT_TERMINAL_PROMPT':'0',
           'GIT_ALLOW_PROTOCOL':'file', 'HTTP_PROXY':'http://127.0.0.1:1',
           'HTTPS_PROXY':'http://127.0.0.1:1', 'ALL_PROXY':'http://127.0.0.1:1'}
    (root/'tmp').mkdir()
    run(env,project,['git','init','-q','-b','main'])
    run(env,project,[str(BIN),'bootstrap','-non-interactive','--skills-root',str(root/'sources')])
    return env, project
def run(env,cwd,args):
    home=pathlib.Path(env['HOME'])
    assert ROOT in home.parents and home.name=='home' and home.stat().st_mode & 0o777 == 0o700
    log('HOME CHECK: private, mode 0700; cwd='+str(cwd))
    log('$ '+shlex.join(args))
    r=subprocess.run(args,cwd=cwd,env=env,text=True,capture_output=True,timeout=90)
    log('stdout:\n'+r.stdout+'stderr:\n'+r.stderr+'exit='+str(r.returncode))
    return r
def cli(env,p,*args): return run(env,p,[str(BIN),*args])
def inventory(p):
    result={}
    if not p.exists(): return result
    for x in sorted(p.rglob('*')):
        rel=str(x.relative_to(p))
        if x.is_symlink(): result[rel]={'link':os.readlink(x),'resolves':x.exists()}
        elif x.is_file(): result[rel]={'sha256':hashlib.sha256(x.read_bytes()).hexdigest(),'bytes':x.stat().st_size}
        else: result[rel]={'directory':True}
    return result
def snap(label,p): log(label+' '+str(p)+'\n'+json.dumps(inventory(p),indent=2))
def skill(p,manifest=True):
    write(p/'SKILL.md','---\nname: sample\ndescription: A local context skill\n---\n# Sample\nLocal guidance.\n')
    if manifest: write(p/'agent-skill.json',json.dumps({'schema_version':4,'capabilities':{},'commands':{},'dependencies':{'skills':{}}}))
def skillfile(p):
    write(p/'Skillfile.json',json.dumps({'schema_version':2,'agents':['codex_cli','claude_code'],'sources':{'local':{'path':'./agents'}},'skills':[{'from':'local','directory':'skills','include':['*']}]}))
def context(p,name):
    write(p/'agent-context.json',json.dumps({'schema_version':1,'name':name,'version':'1.0.0','context':{'modules':[{'path':'a.md'}]}}))
    write(p/'context/a.md',name+' context\n')

env,p=new('A'); skill(p/'agents/skills/sample'); skillfile(p)
cli(env,p,'project','resolve','.')
cli(env,p,'install','.')
write(p/'.agents/skills/unmanaged-skill/SKILL.md','Manual skill\n')
write(p/'.agents/skills/unmanaged-notes/notes.md','Only copy of notes\n')
write(p/'.agents/skills/loose.txt','Only copy of loose text\n')
(p/'.codex/skills/manual-link').symlink_to('../../.agents/skills/unmanaged-skill')
snap('A BEFORE managed',p/'.agents'); snap('A BEFORE native',p/'.codex')
cli(env,p,'install','.')
snap('A AFTER managed',p/'.agents'); snap('A AFTER native',p/'.codex')

env,p=new('B'); skill(p/'agents/skills/sample',False); skillfile(p)
cli(env,p,'skill','check','./agents/skills/sample')
cli(env,p,'project','resolve','.')
cli(env,p,'install','.')
snap('B rejection project output',p/'.agents')
skill(p/'agents/skills/sample',True)
cli(env,p,'skill','check','./agents/skills/sample')
cli(env,p,'project','refresh','.')
cli(env,p,'install','.')

env,p=new('C-project'); skill(p/'agents/skills/sample'); skillfile(p)
write(p/'.agents/.instructions/hand.md','Only copy of hand-authored instructions\n')
write(p/'.agents/README.md','Manual managed-root note\n')
snap('C project BEFORE',p/'.agents'); cli(env,p,'project','resolve','.'); cli(env,p,'install','.'); snap('C project AFTER',p/'.agents')

env,p=new('D-E'); context(p/'alpha','alpha'); context(p/'beta','beta')
snap('D BEFORE',pathlib.Path(env['HOME']))
cli(env,p,'profile','install','./alpha')
cli(env,p,'profile','list')
snap('D AFTER',pathlib.Path(env['HOME']))
cli(env,p,'profile','install','./beta')
cli(env,p,'profile','list')
home=pathlib.Path(env['HOME'])
# Remove only scratch fixture marker, making Claude entrypoint unmanaged.
markers=[home/'.claude/.agent-environment.json']
log('E marker candidates: '+repr([str(x) for x in markers]))
for x in markers:
    if x.is_file(): x.unlink()
write(home/'.claude/CLAUDE.md','Only copy of manual Claude instructions\n')
snap('E BEFORE',home)
cli(env,p,'profile','use','beta')
cli(env,p,'profile','list'); snap('E AFTER failed use',home)
cli(env,p,'profile','use','alpha'); snap('E AFTER recovery attempt',home)
cli(env,p,'profile','use','beta','--takeover'); snap('C native AFTER takeover',home)
cli(env,p,'profile','list')
