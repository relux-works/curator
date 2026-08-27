# TASK-260720-1zl1cj review verdict cycle 2

## Verdict

Changes requested; route to implementation rework.

## Correctness finding

The longest-existing-prefix rework still does not produce one canonical first-use home identity on Windows. internal/managerlock/identity.go lines 74-89 resolves the existing prefix and appends every missing component verbatim. Windows filepath.EvalSymlinks canonicalizes case only for components that exist. Therefore two independent managers created before the home exists from case-variant configured spellings can retain different Manager.home values even though ordinary Windows path semantics make the created directory and lock root the same physical location. managerlock.go line 44 splits process order state by those strings, and lines 293-295 hash them into distinct home-lock filenames. The two processes can consequently acquire different files concurrently instead of one exclusive manager-home lock. The portable same-spelling regression does not cover this Windows-specific alias case.

Required rework: make first-use identity and manager-home exclusion stable for case aliases on case-insensitive Windows without breaking case-sensitive filesystems; add a Windows subprocess regression that constructs case-variant managers while the home is absent, proves the same physical home identity or equivalent single lock path after creation, proves contention before release and acquisition after release, and verifies process-local order state is not split. Preserve the existing symlinked-prefix regression. Run the Windows subprocess suite on a native Windows runner when available.

## Passing evidence

Independent review passed go test -race -cover ./internal/managerlock -count=1 -v at 82.2 percent coverage, go test -race ./... -count=1, make check, go vet ./internal/managerlock, gofmt, git diff --check, no-staging verification, and Linux amd64 plus Windows amd64 managerlock test compilation. Native Windows runtime was unavailable on Darwin. No product code was modified during review.