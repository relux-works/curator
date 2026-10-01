import os, pathlib, subprocess, tempfile, json
root=pathlib.Path(tempfile.mkdtemp(prefix='TASK-261001-2yvag1-launch-'))
bin=root/'bin';bin.mkdir();(bin/'curator').symlink_to('/tmp/TASK-261001-2yvag1-curator')
(bin/'muse').write_text('#!/bin/sh\nif [ "$1" = "--version" ]; then echo 1.4.1-R4503.1; else /usr/bin/env; fi\n');(bin/'muse').chmod(0o755)
env=os.environ.copy()
env.update(HOME=str(root/'operator'),CURATOR_CONFIG=str(root/'manager'/'config.json'),XDG_CONFIG_HOME=str(root/'native-config'),CODEX_HOME=str(root/'codex'),CLAUDE_CONFIG_DIR=str(root/'claude'),PI_CODING_AGENT_DIR=str(root/'pi'),PATH=str(bin)+os.pathsep+env['PATH'])
for name in ['operator','manager','native-config/muse','codex','claude','pi/agent','profile/context']:(root/name).mkdir(parents=True,exist_ok=True)
(root/'manager/config.json').write_text(json.dumps({'schema_version':2,'skills_root':str(root/'skills'),'projects':{}}))
(root/'native-config/muse/auth.json').write_text('synthetic-native\n')
(root/'codex/auth.json').write_text('synthetic-codex\n');(root/'pi/agent/auth.json').write_text('synthetic-pi\n')
(root/'profile/agent-context.json').write_text(json.dumps({'schema_version':1,'name':'acme','version':'1.0.0','context':{'modules':[{'path':'a.md'}]}}));(root/'profile/context/a.md').write_text('hello\n')
for parent,dirs,files in os.walk(root/'profile'):
 os.chmod(parent,0o700)
 for f in files:os.chmod(pathlib.Path(parent)/f,0o600)
commands=[['curator','profile','install',str(root/'profile')],['curator','env','resolve','muse','--repair','--format','json'],['curator','run','muse','--','exec','synthetic']]
for cmd in commands:
 result=subprocess.run(cmd,env=env,capture_output=True,text=True)
 print('COMMAND',json.dumps(cmd),'EXIT',result.returncode)
 print(result.stdout);print(result.stderr)
 if result.returncode and cmd[1]!='run':raise SystemExit(result.returncode)
 if cmd[1]=='run':raise SystemExit(result.returncode)
