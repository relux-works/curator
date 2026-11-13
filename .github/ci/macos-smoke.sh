#!/usr/bin/env bash
# Native smoke keeps exact PASS witnesses, so a renamed/filtered/skipped test fails.
set -euo pipefail
out=.temp/ci-evidence/macos-smoke
mkdir -p "$out"
bash .github/ci/isolated-test.sh go test -count=1 -timeout=3m -json \
  ./internal/envprofile ./internal/install ./internal/transaction \
  -run '^Test(ClaudeSeedPreparedRename|DefaultTransportInstallScopes|ExchangePreparationContract)$' \
  | tee "$out/tests.json"
python3 .github/ci/required-tests.py "$out/tests.json" \
  TestClaudeSeedPreparedRename TestDefaultTransportInstallScopes TestExchangePreparationContract
python3 .github/ci/required-tests.py "$out/tests.json" \
  TestDefaultTransportInstallScopes/project/selected \
  TestDefaultTransportInstallScopes/project/pinned-agent \
  TestDefaultTransportInstallScopes/project/wrong-pin \
  TestDefaultTransportInstallScopes/global/selected \
  TestDefaultTransportInstallScopes/global/pinned-agent \
  TestDefaultTransportInstallScopes/global/wrong-pin \
  TestClaudeSeedPreparedRename/absent TestClaudeSeedPreparedRename/regular \
  TestClaudeSeedPreparedRename/directory TestClaudeSeedPreparedRename/invalid-path \
  TestExchangePreparationContract/valid TestExchangePreparationContract/absent \
  TestExchangePreparationContract/removal TestExchangePreparationContract/generation
