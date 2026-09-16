# STORY-260916-1ll22r: absence-vs-read-failure-collapse

## Description
Security-class correctness gap from the launcher/migration campaign: environments §8.4 requires that an absent file and a failed read stay different facts and that a read or parse error is never downgraded to absence; the campaign found the two collapsed at five sites over three review cycles, and the producer stated there is no structural barrier against the next one. Make the discipline structural.

## Scope
curator internal/* read paths (os.IsNotExist branches, seed/marker/lock/passthrough readers), a shared helper, a lint or test guard, conformance vectors

## Acceptance Criteria
Inventory of every read site that maps an error to absence with a verdict per site; a shared classification helper used at every site; a guard (analyzer, lint rule or test) that fails the build when an os.IsNotExist branch swallows another error class; vectors exercising unreadable-but-present for markers, seeds, locks and passthrough entries
