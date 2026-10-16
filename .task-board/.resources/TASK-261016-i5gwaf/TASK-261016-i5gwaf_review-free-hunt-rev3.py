from pathlib import Path
import shutil,subprocess,os,json
root=Path.cwd(); out=root/'.temp/review-muse-r3/free-hunt'; out.mkdir()
source=root/'.research/muse148'; spec=root/'.temp/TASK-261016-i5gwaf/spec-main'; python=root/'.temp/TASK-261016-i5gwaf/reference-venv/bin/python'
# Challenge the real reference CLI with physical and identity corruption.
for shape in ['regular','directory-link','metadata-link','root-link','missing-root','changed-root','extra-file']:
 dest=out/shape; shutil.copytree(source/'testdata/snapshot',dest)
 file=dest/'home/config/muse/AGENTS.md'; file.chmod(0o600)
 if shape=='directory-link':
  real=out/'external-config'; shutil.move(dest/'home/config',real); (dest/'home/config').symlink_to(real,target_is_directory=True)
 elif shape=='metadata-link':
  real=out/'external-manifest.json'; shutil.move(dest/'manifest.json',real); (dest/'manifest.json').symlink_to(real)
 elif shape=='root-link':
  real=out/'external-bundle'; dest.rename(real); dest.symlink_to(real,target_is_directory=True)
 elif shape=='missing-root': file.unlink()
 elif shape=='changed-root': file.write_text('MUSE148_ALTERED\n')
 elif shape=='extra-file': (dest/'home/extra').write_text('extra\n')
 before=file.read_bytes() if file.exists() else None
 run=subprocess.run([str(python),'-B',str(source/'reference-snapshot-check.py'),str(spec),str(dest)],capture_output=True,text=True)
 (out/(shape+'.log')).write_text(run.stdout+run.stderr)
 expected=0 if shape=='regular' else 1
 assert run.returncode==expected,(shape,run.returncode)
 if shape in ['directory-link','metadata-link','root-link']: assert 'ValueError: physical-entry-type' in run.stderr
 assert (file.read_bytes() if file.exists() else None)==before
 print(shape,run.returncode,run.stderr.strip().splitlines()[-1:] or 'accepted')
