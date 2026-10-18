# Claim: malformed configuration cannot preempt help; mixed help positions and
# concrete compose profile operands preserve dedicated output and persistent state.
import hashlib, os, pathlib, subprocess, tempfile
binary = pathlib.Path('.temp/review-dkgv1p/curator').resolve()
with tempfile.TemporaryDirectory(dir='.temp/review-dkgv1p') as tmp:
    root = pathlib.Path(tmp).resolve()
    for name in ('cwd','home','curator','xdg','tmp'):
        (root/name).mkdir()
    cfg=root/'curator'/'config.json'
    cfg.write_text('{ malformed configuration')
    env={k:v for k,v in os.environ.items() if not k.startswith(('CURATOR_','XDG_'))}
    env.update(HOME=str(root/'home'),CURATOR_HOME=str(root/'curator'),CURATOR_CONFIG=str(cfg),XDG_CONFIG_HOME=str(root/'xdg'),TMPDIR=str(root/'tmp'),TMP=str(root/'tmp'),TEMP=str(root/'tmp'),GIT_CONFIG_GLOBAL=os.devnull,GIT_CONFIG_NOSYSTEM='1')
    def snap():
        return {str(p.relative_to(root)):(p.lstat().st_mode,p.lstat().st_size,hashlib.sha256(p.read_bytes()).hexdigest() if p.is_file() else None) for p in root.rglob('*')}
    cases=[(['profile','compose','example','add'],'profile compose <profile> add'),(['env','config','set'],'env config set'),(['config','build-https','login'],'config build-https login'),(['init'],'init')]
    n=0
    for path,topic in cases:
        baseline=subprocess.run([str(binary),'help',*path],cwd=root/'cwd',env=env,capture_output=True)
        assert baseline.returncode==0 and baseline.stderr==b''
        assert ('usage: curator '+topic).encode() in baseline.stdout
        for pos in range(1,len(path)+1):
            args=path[:pos]+['--help','-h']+path[pos:]+['--','-h']
            before=snap()
            result=subprocess.run([str(binary),*args],cwd=root/'cwd',env=env,capture_output=True)
            assert result.returncode==0 and result.stderr==b'' and result.stdout==baseline.stdout,args
            assert snap()==before,args
            n+=1
    before=snap()
    result=subprocess.run([str(binary),'help','unknown-topic'],cwd=root/'cwd',env=env,capture_output=True)
    assert result.returncode==2 and result.stdout==b'' and b'unknown help topic' in result.stderr
    assert snap()==before
    print(f'PASS: {n} mixed-position cases with malformed config; unknown topic exit 2; snapshots unchanged')
