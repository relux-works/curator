from pathlib import Path
import shutil,subprocess,os
root=Path('/tmp/s9-review2')
mutants=[('activation_plain_only','    resolved.store(active_key_path(home), staged)\n    staged_path.unlink()', '    if b"ENCRYPTED PRIVATE KEY" in staged_path.read_bytes():\n        resolved.store(active_key_path(home), staged)\n        staged_path.unlink()\n    else:\n        os.replace(staged_path, active_key_path(home))', 'test_activation_encrypts_plain_staged_key'),('empty_load_only','    return FileKeyProvider(passphrase=key_passphrase_from_env())','    return FileKeyProvider(passphrase=None if os.environ.get(KEY_PASSPHRASE_ENV) == "" else key_passphrase_from_env())','test_empty_passphrase_at_load_time_fails_closed')]
for name,old,new,test in mutants:
 tree=root/name; shutil.copytree(root/'source',tree,ignore=shutil.ignore_patterns('__pycache__','.pytest_cache','.mypy_cache','*.egg-info'))
 p=tree/'src/csk_registry/keys.py'; text=p.read_text(); assert old in text; p.write_text(text.replace(old,new,1))
 r=subprocess.run([str(root/'venv/bin/python'),'-m','pytest','-q','tests/test_key_passphrase.py::'+test],cwd=tree,text=True,capture_output=True)
 (root/(name+'.log')).write_text(r.stdout+r.stderr)
 print(name,'exit',r.returncode); print(r.stdout)
 assert r.returncode==1
print('Narrowing mutants caught: 2/2 selected; not exhaustive mutation coverage')
