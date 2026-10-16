"""Task-local narrowing controls: generate Go overlays and run Go behavioral tests.
Run from the candidate repository root. Writes only under .temp/.
"""
import json
from pathlib import Path
import subprocess
import sys

root = Path.cwd()
base = root / '.temp/TASK-261016-hzbo83/review-controls'
base.mkdir(parents=True, exist_ok=True)
rows = [
    ('M1-auto-isolated', 'credential_policy.go', 'p.nativeStore != "file" {', 'p.nativeStore != "file" && p.nativeStore != "auto" {', 'TestCredentialProjectionIsolationRefusals/auto'),
    ('M2-muse-file-admitted', 'credential_policy.go', 'if selection == envregistry.CredentialStoreNative {', 'if selection == envregistry.CredentialStoreNative || adapter.ID == envregistry.Muse && selection == envregistry.CredentialStoreFile {', 'TestCredentialStorePolicyActivation/muse/file'),
    ('M3-permission-defaults-file', 'managed.go', 'state, err := readRegularFile(path)\n\tif err != nil {', 'state, err := readRegularFile(path)\n if errors.Is(err, os.ErrPermission) { return "file", nil }\n\tif err != nil {', 'TestCredentialStorePolicyNativeInspection/permission'),
    ('M4-keyring-resurrects-file', 'credential_policy.go', 'if p.nativeStore == "keyring" {', 'if p.nativeStore == "keyring" && p.selection == envregistry.CredentialStoreKeychain {', 'TestCodexKeyringAmbient/keyring'),
    ('M5-shared-second-observation', 'credential_policy.go', 'cached.environment == adapter.ID {', 'cached.environment == adapter.ID && (cached.policy == nil || cached.policy.isolation != envregistry.IsolationShared) {', 'TestCredentialProjectionSingleObservation/shared'),
    ('M6-error-reobservation', 'credential_policy.go', 'cached.environment == adapter.ID {', 'cached.environment == adapter.ID && cached.err == nil {', 'TestCredentialProjectionSingleObservation'),
    ('M7-pi-subrequest-reuses-policy', 'identity_migration.go', 'rr.credentialPolicy = nil', 'if adapter.ID != envregistry.Pi { rr.credentialPolicy = nil }', 'TestCredentialProjectionIdentitySubrequest/isolated'),
]
results = []
for name, filename, old, new, test in rows:
    source = root / 'internal/envprofile' / filename
    text = source.read_text()
    if text.count(old) != 1:
        raise SystemExit(f'{name}: expected exactly one plant location; control not applied')
    mutated = base / (name + '.go')
    mutated.write_text(text.replace(old, new))
    overlay = base / (name + '.json')
    overlay.write_text(json.dumps({'Replace': {str(source): str(mutated)}}))
    log = base / (name + '.log')
    command = ['go', 'test', '-overlay', str(overlay), './internal/envprofile', '-run', '^' + test + '$', '-count=1']
    with log.open('w') as output:
        process = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT, check=False)
    killed = process.returncode == 1 and ('--- FAIL: ' + test) in log.read_text()
    result = {'mutant': name, 'exit_code': process.returncode, 'named_failure': test, 'killed': killed}
    results.append(result)
    print(json.dumps(result), flush=True)
(base / 'results.json').write_text(json.dumps(results, indent=2))
sys.exit(0 if all(result['killed'] for result in results) else 1)
