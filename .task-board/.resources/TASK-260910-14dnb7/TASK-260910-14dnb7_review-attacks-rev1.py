from pathlib import Path
import os, subprocess
root=Path('/tmp/csk-review-14dnb7')
app=root/'src/csk_registry/app.py'
original=app.read_text()
env=dict(os.environ, CURATOR_CONFORMANCE_ROOT='/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1')
python='/tmp/csk-review-14dnb7-venv/bin/python'
mutants={
 'records_cursor_missing': original.replace('return {"records": found, "next_cursor": next_cursor, "boundary": boundary_snapshot}', 'return {"records": found, "next_cursor": next_cursor, **({} if cursor else {"boundary": boundary_snapshot})}'),
 'log_cursor_latest': original.replace('return {"entries": encoded, "next_cursor": next_cursor, "boundary": boundary_snapshot}', 'return {"entries": encoded, "next_cursor": next_cursor, "boundary": build_snapshot(store, signing_key) if cursor else boundary_snapshot}'),
 'log_bad_hash': original.replace('"entry_hash": entry.entry_hash', '"entry_hash": "invalid"'),
}
for name,source in mutants.items():
 try:
  app.write_text(source)
  args=['-m','pytest','-q']
  if name != 'log_bad_hash': args+=['tests/test_registry.py','-k','pages_carry_byte_identical_boundary']
  result=subprocess.run([python,*args],cwd=root,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  Path('/tmp/csk-review-14dnb7-'+name+'.log').write_text(result.stdout+f'\nexit={result.returncode}\n')
  print(name, result.returncode, result.stdout[-350:], flush=True)
 finally: app.write_text(original)
# Extend the existing real rotation entry-point test with the missing assertion.
test=root/'tests/test_registry.py'
original_test=test.read_text()
try:
 test.write_text(original_test.replace('assert continued.status_code == 200', 'assert continued.status_code == 200\n    assert continued.json()["boundary"] == first.json()["boundary"]'))
 result=subprocess.run([python,'-m','pytest','-q','tests/test_registry.py','-k','staged_key_rotation'],cwd=root,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
 Path('/tmp/csk-review-14dnb7-rotation.log').write_text(result.stdout+f'\nexit={result.returncode}\n')
 print('rotation',result.returncode,result.stdout[-1500:],flush=True)
finally: test.write_text(original_test)
