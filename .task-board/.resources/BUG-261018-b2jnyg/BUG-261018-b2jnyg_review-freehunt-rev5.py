# Reproduce the reviewer's additional valid-number/string preservation probe.
import json
import pathlib
import subprocess
p = pathlib.Path('.temp/review-b2jnyg')
p.mkdir(parents=True, exist_ok=True)
original = pathlib.Path('internal/envprofile/claude_seed_test.go').resolve()
source = original.read_text()
needle = '"unknown":{"text":"tool state","enabled":false}'
replacement = r'"unknown":{"text":"<>&\u0000\ud83d\ude00","enabled":false},"extraNumbers":[-0.0000,0e-0,1.23000e+0004,-999999999999999999999999999999,1E-999]'
assert source.count(needle) == 1
(p / 'freehunt_test.go').write_text(source.replace(needle, replacement))
(p / 'freehunt-overlay.json').write_text(json.dumps({'Replace': {str(original): str((p / 'freehunt_test.go').resolve())}}))
raise SystemExit(subprocess.call(['go', 'test', '-overlay', str(p / 'freehunt-overlay.json'), './internal/envprofile', '-count=1', '-timeout=3m', '-v', '-run', '^TestClaudeRepairPreservesNumberText$/^(referenced|monolithic)$/^new-launch$']))
