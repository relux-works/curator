# TASK-260728-20ao7p — Windows gate fix 1: the black-box found a real Windows defect (THE ONLY CURRENT INSTRUCTION, with 20ao7p-brief.md)

Rev1 (tree adc6bc61) passed everywhere except Test (windows-latest), run 36796244453. TestNativeBlackboxExternalBuildLifecycle fails at
native_blackbox_test.go:158:
- shim exit=1;
- stderr: `'"…\home\external-build-cache\artifacts\<key>\artifact"' is not recognized as an internal or external command`.

The shim for an external-repository build on Windows invokes an artifact path WITHOUT `.exe` (`artifact`), so cmd.exe cannot run it.
This is exactly what the black-box exists to catch, and it is very likely a PRODUCTION bug in the external-build pipeline on Windows.
1. Find where the external-build artifact name and the shim target are produced: internal/buildrepo pipeline, internal/install
   external.go, and the shim writer. Compare with the local go-v1 build path, which works on Windows. Fix the production code so the
   built artifact and the shim target carry the platform executable suffix on Windows (or whatever the local build path does).
   Then check that the receipt/marker and the cache key stay consistent with rc.13; cite the spec clause on the artifact name if one
   exists.
2. The same log warns: `build cache sweep skipped: untrusted cache provenance: …\home DACL is not protected from inheritance`. The test
   HOME must be created through the product's private-directory helper on Windows, so the run is representative. Fix the fixture.
3. Keep the black-box assertions. Do not skip on Windows. Add a focused unit test at the production seam for the Windows artifact name
   (it can run on all OSes using the GOOS-specific suffix logic).
4. Run `GOOS=windows go vet ./...` (the touched packages) and `go test ./cmd/curator -run NativeBlackbox`, plus the focused test, with
   real exit codes. Report the root cause precisely in the results.
Set status development, update the results, run `task-board handoff TASK-260728-20ao7p --role developer`, then END YOUR TURN. No
CHANGELOG/LOGBOOK edit: put the entry text in the results.
