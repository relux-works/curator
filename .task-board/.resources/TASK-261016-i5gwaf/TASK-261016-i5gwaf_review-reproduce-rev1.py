"""Reproduce both review findings using exact CR revision 1 candidate blobs.
Run from the repository root with SPEC_PYTHON and PINNED_SPEC set.
All generated files remain in a fresh ignored .temp directory.
"""
from pathlib import Path
import os
import shutil
import subprocess
import tempfile

candidate = '0f67bfb2d484e4e1f32711d4d8b76012aca05680'
source = Path('.research/muse148')
paths = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', candidate, str(source)], text=True).splitlines()
for name in paths:
    expected = subprocess.check_output(['git', 'show', candidate + ':' + name])
    assert Path(name).read_bytes() == expected, 'candidate mismatch: ' + name
scratch = Path(tempfile.mkdtemp(prefix='muse148-review-', dir='.temp')).resolve().relative_to(Path.cwd())
checker = scratch / 'checker'
shutil.copytree(source, checker)
(checker / 'review_attack_test.go').write_text('''package muse148
import "testing"
// Contradictory loader fields cannot attest a successful zero-skill result.
func TestReviewContradictoryObservation(t *testing.T) {
 for _, log := range []string{
  `event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=1`,
  `event="skills.catalog_load" outcome="failed" outcome="ready" user_skills=0`,
 } {
  if checkCounts(log,"skills.catalog_load",map[string]int{"user_skills":0}) == nil {
   t.Errorf("contradictory record admitted: %s",log)
  }
 }
}
''')
def run(label, args, expected):
    result = subprocess.run(args, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    (scratch / (label + '.log')).write_text(result.stdout)
    print(label + ': exit=' + str(result.returncode) + '; contract expects=' + str(expected), flush=True)
    print(result.stdout, end='', flush=True)
    return result.returncode

run('checker', ['go', 'test', './'+str(checker), '-run', '^Test(ObservationChecker|ReviewContradictoryObservation)$', '-count=1', '-v'], 0)
for shape in ('positive', 'listed-link', 'unlisted-dangling-link'):
    package = scratch / shape
    shutil.copytree(source, package)
    bundle = package / 'testdata/snapshot'
    root = bundle / 'home/config/muse/AGENTS.md'
    root.chmod(0o600)
    if shape == 'listed-link':
        target = scratch / 'outside-root.md'
        shutil.copyfile(root, target)
        target.chmod(0o600)
        root.unlink()
        root.symlink_to(target.resolve())
    if shape == 'unlisted-dangling-link':
        (bundle / 'home/hidden-link').symlink_to('missing-target')
    if shape == 'listed-link':
        run('frozen-go-listed-link', ['go', 'test', './'+str(package), '-run', '^TestFrozenSnapshot$', '-count=1', '-v'], 1)
    run('reference-'+shape, [os.environ['SPEC_PYTHON'], '-B', str(source/'reference-snapshot-check.py'), os.environ['PINNED_SPEC'], str(bundle)], 0 if shape == 'positive' else 1)
print('Review findings: contradictory loader fields admitted; linked and unlisted-link snapshot inputs admitted.')
