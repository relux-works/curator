import os, pathlib, subprocess, sys, runpy, tempfile
root=pathlib.Path('/tmp/csk-review-r2-mutants')
os.chdir(root)
sys.path.insert(0,str(root/'src'))
os.environ['CURATOR_CONFORMANCE_ROOT']='/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1'
app=root/'src/csk_registry/app.py'
original=app.read_text()
mutants=[
 ('records-resign-cursor-only', 'return {"records": found, "next_cursor": next_cursor, "boundary": boundary_snapshot}', 'return {"records": found, "next_cursor": next_cursor, "boundary": build_snapshot(store, signing_key, boundary=boundary) if cursor else boundary_snapshot}', 'tests/test_registry.py::test_rotation_overlap_keeps_chain_boundary_on_both_endpoints'),
 ('log-resign-cursor-only', 'return {"entries": encoded, "next_cursor": next_cursor, "boundary": boundary_snapshot}', 'return {"entries": encoded, "next_cursor": next_cursor, "boundary": build_snapshot(store, signing_key, boundary=boundary) if cursor else boundary_snapshot}', 'tests/test_registry.py::test_rotation_overlap_keeps_chain_boundary_on_both_endpoints'),
 ('snapshot-check-records-only', 'if not any(verify_signed(public_key, snapshot) for public_key in verification_keys):', 'if endpoint == "records" and not any(verify_signed(public_key, snapshot) for public_key in verification_keys):', 'tests/test_registry.py::test_rotation_overlap_keeps_chain_boundary_on_both_endpoints'),
 ('malformed-log-hash-only', '"entries": encoded, "next_cursor": next_cursor, "boundary": boundary_snapshot', '"entries": [{**entry, "entry_hash": "invalid"} for entry in encoded], "next_cursor": next_cursor, "boundary": boundary_snapshot', 'tests/test_protocol_conformance.py::test_shared_service_page_boundary_vectors'),
]
try:
 for name, old, new, test in mutants:
  assert original.count(old)==1,(name,original.count(old))
  app.write_text(original.replace(old,new))
  p=subprocess.run([sys.executable,'-m','pytest','-q',test],capture_output=True,text=True)
  print('\nMUTANT',name,'EXIT',p.returncode,flush=True)
  print(p.stdout, p.stderr,flush=True)
  assert p.returncode==1
finally:
 app.write_text(original)
assert app.read_text()==original
# Independently execute the committed rotation scenario with instrumented cursor lengths.
import csk_registry.app as application
encode=application._encode_cursor
lengths={}
def measured(*args,**kwargs):
 result=encode(*args,**kwargs)
 lengths.setdefault(kwargs['endpoint'],[]).append(len(result))
 return result
application._encode_cursor=measured
ns=runpy.run_path(str(root/'tests/test_registry.py'))
with tempfile.TemporaryDirectory() as tmp:
 ns['test_rotation_overlap_keeps_chain_boundary_on_both_endpoints'](pathlib.Path(tmp))
print('Independent rotation replay PASS; cursor lengths',lengths)
ns2=runpy.run_path(str(root/'tests/test_protocol_conformance.py'))
with tempfile.TemporaryDirectory() as tmp:
 ns2['test_shared_service_log_harness_rejects_malformed_entry_hash'](pathlib.Path(tmp))
print('Independent malformed served hash rejection PASS')
