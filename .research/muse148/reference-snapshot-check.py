import os, stat, sys
from pathlib import Path
spec=Path(sys.argv[1]).resolve()
sys.path.insert(0,str(spec/'tools'))
from test_validate import validate
import managed_home_snapshots as snapshots
root=Path(sys.argv[2]).absolute()
# check_bundle accepts an in-memory inventory, not a filesystem safety claim.
# Classify every physical entry without following links before collecting bytes;
# dangling and unlisted links must not disappear from that inventory.
def inventory(directory):
    if not stat.S_ISDIR(directory.lstat().st_mode):
        raise ValueError('physical-entry-type')
    content = {}
    with os.scandir(directory) as entries:
        for entry in entries:
            path = Path(entry.path)
            mode = entry.stat(follow_symlinks=False).st_mode
            if stat.S_ISDIR(mode):
                content.update(inventory(path))
            elif stat.S_ISREG(mode):
                content[path.relative_to(root).as_posix()] = (path.read_bytes(), f'{mode & 0o777:04o}')
            else:
                raise ValueError('physical-entry-type')
    return content

if not stat.S_ISDIR(root.lstat().st_mode):
    raise ValueError('physical-entry-type')
for name in ('manifest.json', 'launch.json'):
    if not stat.S_ISREG((root/name).lstat().st_mode):
        raise ValueError('physical-entry-type')
content=inventory(root/'home')
manifest=validate.load_json(root/'manifest.json')
launch=validate.load_json(root/'launch.json')
if os.environ.get('MUSE148_SNAPSHOT_PLANT')=='extra': content['home/extra']=(b'extra','0600')
snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
print('Reference check_bundle accepted exact frozen metadata and complete regular-file closure')
