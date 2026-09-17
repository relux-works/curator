import os,pathlib,tempfile,subprocess
b=pathlib.Path(tempfile.mkdtemp(prefix='TASK-260910-3ungjy-e2e-',dir='/tmp')).resolve()
exe='/tmp/TASK-260910-3ungjy-curator'
env=dict(os.environ,CURATOR_CONFIG=str(b/'home/config.json'))
p=b/'project'; (p/'.agents').mkdir(parents=True)
for shell in ['sh','bash','powershell']:
 f=p/'.agents'/('env.ps1' if shell=='powershell' else 'env.sh')
 f.write_text('$env:REVIEW_MARKER="yes"\n' if shell=='powershell' else 'export REVIEW_MARKER=yes\n')
 hook=b/('hook.ps1' if shell=='powershell' else 'hook.sh')
 r=subprocess.run([exe,'shell-init',shell if shell=='powershell' else 'bash'],env=env,text=True,capture_output=True); assert r.returncode==0,(r.stdout,r.stderr)
 hook.write_text(r.stdout)
 for action in ['approve','revoke']:
  r=subprocess.run([exe,'hook',action,str(f)],env=env,text=True,capture_output=True); print(shell,action,'exit',r.returncode,r.stdout.strip(),r.stderr.strip()); assert r.returncode==0
  if shell=='powershell':
   argv=['/tmp/pwsh/app/pwsh','-NoProfile','-NonInteractive','-Command',f'. "{hook}"; Write-Output "marker=$env:REVIEW_MARKER"']
  else: argv=[shell,'-c',f'. "{hook}"; printf "marker=%s\\n" "$REVIEW_MARKER"']
  r=subprocess.run(argv,cwd=p,env=env,text=True,capture_output=True,timeout=45)
  print(shell,'activation after',action,'exit',r.returncode,'stdout',repr(r.stdout),'stderr',repr(r.stderr))
  assert r.returncode==0 and 'marker=yes' in r.stdout
  assert ('shell_hook_env_unapproved' in r.stderr)==(action=='revoke')
  if action=='approve': assert not r.stderr
print('E2E passed 6/6 activation checks')
