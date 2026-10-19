# BUG-261018-dbeghc — successor publication handoff


## Publication after converge — successor handoff recovery

The predecessor attached its r5 evidence but issued no developer handoff. This successor changes no source, test or documentation files. Base remains `892dab098aef1497df9ccacdb173faaf4f8b8418`; candidate tree remains `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. All five task-file SHA256 values match the predecessor evidence and those files remain byte-identical to signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`. The earlier README removal is already carried; README and CHANGELOG equal trunk. No refresh-specific refusal is reported, so no refresh was run. No commit or branch mutation was made.

This successor directly reran the six permission suites (including all six narrowing controls) with `go test -count=1 -timeout=4m -json -run '^TestManaged(SeedPreservesNativePermissions|HomeRepairTightensDirectories|AuthoredFilesArePrivate|SeedRefusalsPreserved|FallbackPermissions|ModesNarrowingMutants)$' ./internal/envprofile`, exit 0; `go build ./...`, exit 0; README/CHANGELOG equality, index equality, diff whitespace and changed-file gofmt checks, exit 0 each. Exact identities, command and test verdicts are attached as `BUG-261018-dbeghc_successor-checks.json`. Earlier conformance and CLI verification remains the predecessor's evidence, not a claim that this successor reran it; interrupted and failing diagnostics remain explicitly red above.

Product coverage remains **4 of 5 AC rows driven** with the production call sites, out-of-contract declarations and six-mutant table above. No reviewer tests removed or renamed. No new platform claim. Full local suite and full hosted gate on this combined tree are not claimed: r5 assigns configured validation after handoff. Existing checklist row 5 is left unchecked; no new post-turn-validation checklist item is added. This successor will invoke the requested developer handoff after attaching evidence, then end its turn. Any handoff refusal is an actual refusal, not review acceptance.

Logbook: successor confirms the carried tree is unchanged and closes the predecessor's missing-handoff step. No shared runner, service, process, credential or control-root file was changed directly.
