import os, pathlib, subprocess, json, tempfile
root = pathlib.Path('snapshot').resolve()
workflow = root / '.github/workflows/ci.yml'
data = workflow.read_text()
mutant = data.replace('          actual="$(go version)"', '          export GOCACHE="$RUNNER_TEMP/shared-build" GOMODCACHE="$RUNNER_TEMP/shared-mod" GOPATH="$RUNNER_TEMP/shared-path"\n          actual="$(go version)"', 1)
assert mutant != data
path = pathlib.Path('shell-override.yml').resolve()
path.write_text(mutant)
env = dict(os.environ, CI_WORKFLOW_UNDER_TEST=str(path))
with open('shell-override-suite.log', 'w') as log:
    result = subprocess.run(['go','test','-count=1','-timeout=3m','-v','./tools/ciworkflow/'], cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
print('Token-preserving shell override full suite exit:', result.returncode)
with tempfile.TemporaryDirectory() as temp:
    e = dict(os.environ, RUNNER_TEMP=temp, GITHUB_ENV=temp+'/env', GOTOOLCHAIN='local', GOENV='off')
    subprocess.run(['bash',str(root/'.github/ci/isolate-go-state.sh')],env=e,check=True)
    goroot = subprocess.check_output(['go', 'env', 'GOROOT'], text=True).strip()
    e['PATH'] = goroot+'/bin:'+e['PATH']
    e.update(dict(line.split('=',1) for line in pathlib.Path(e['GITHUB_ENV']).read_text().splitlines()))
    # Same prepended shell statement as the first Go-consuming workflow step.
    command = 'export GOCACHE="$RUNNER_TEMP/shared-build" GOMODCACHE="$RUNNER_TEMP/shared-mod" GOPATH="$RUNNER_TEMP/shared-path"; go env GOCACHE GOMODCACHE GOPATH'
    result2 = subprocess.run(['bash','-c',command],env=e,text=True,capture_output=True,check=True)
    print('Observed:',result2.stdout.replace(temp, '<TEMP>')); assert result2.stdout.splitlines() == [temp+'/shared-build',temp+'/shared-mod',temp+'/shared-path']
    print('Real Go consumed all three shared overrides: yes')
assert result.returncode == 0, 'Survivor did not reproduce'
