# Logbook — TASK-260906-2x4s7i, cycle 8

## A regex character class is a portability boundary, and nobody was looking at it

Seven cycles attacked this discriminator through the spelling space: more hosts, more schemes, more
drive letters, more slashes. Every one of them stayed inside printable ASCII. The defect that was
still there sat one layer down, in `\s` — a character class whose *membership* differs between
CPython `re`, ECMA-262 and Go RE2, three engines all of which have a claim on a JSON Schema
`pattern` in this repository. Insert U+00A0 and every refusal the last four commits added evaporates
into a valid form-free `path`.

The lesson generalises past this change: when a normative artefact carries a regex, the regex's
*shorthand classes* are part of its meaning, and shorthand classes are the least portable thing in
regular expressions. `\s`, `\w`, `\b`, `\d` under Unicode, and `$`-before-trailing-newline are all
places where two conforming engines answer differently. A spec that pins behaviour with a pattern
and validates it with exactly one engine has not pinned it.

## How to find a class no spelling corpus contains: mutate the class, not the spelling

The mutant that exposed it was `^[^/\s]*:` -> `^[^/]*:`. Zero flips on the published corpus, 160
flips over an independent population. That gap *is* the finding: a mutant with a large measured
behavioural delta and zero corpus flips is a blind spot with a size attached to it. Reporting "this
mutant survives" says almost nothing; reporting "this mutant survives and moves 160 of 412
spellings" says exactly how much the corpus cannot see.

Cheap generalisation: after a spelling sweep comes back clean, sweep the *character classes* the
patterns are built from. It is a small, finite mutant family and it finds a different kind of bug.

## Scope a finding by asking who reads the artefact, then measure it

The engine divergence looked blocking until I checked whether anything actually consumes the schema.
`.github/ci/implementation-coverage.tsv` names no `schemas/` path (0 occurrences) and
`conformance/v1/manifest.json` digests no path under `schemas/` (0 of 2097 string values); the
implementation lanes publish `conformance/v1` as their root, which does not contain the schema set.
So the divergence is latent, and the finding is MINOR rather than MAJOR — reported with the
measurement that makes it MINOR, and with the condition under which it stops being MINOR.

The inverse failure would have been to call it blocking from the regex table alone. Severity is a
function of the consumer, and the consumer is a thing you read, not a thing you assume.

## A measured zero is only as wide as its generator

Cycle 7 published "0 colon-shaped non-drive spellings escaping to `path`" over 3,916 spellings and
was right about those 3,916. The sentence was wider than the generator. This is the standing shape:
an honest measurement wrapped in a sentence the measurement does not support. The fix is mechanical
— state the population next to the zero — and it is worth doing every time, because the next
reviewer will otherwise treat the class as closed and look somewhere else. I did, for six cycles
of artefacts, until I mutated the class instead of reading the verdict.

## Controls have to survive on more than the corpus they were built for

`C-reorder-allof` is meant to be a semantic no-op. Surviving the 112-entry gate corpus proves very
little, because most things survive that corpus. Running it over the 412-spelling population and
measuring 0 flips there is what makes it a control. A harness control that is only checked against
the weak signal is not a control.

## Four clean cycles is itself a result

Cycles 1–4 each found a real defect. Cycles 5–8 found nothing blocking; this cycle's three new
findings are two prose corrections and one latent bound. A review loop that cannot terminate on
"nothing blocking" spends its budget re-deriving settled behaviour, and each extra cycle has a lower
prior of finding anything. The termination signal is not "a cycle found nothing" — it is several
consecutive cycles finding nothing *while attacking different surfaces*, which is now the case:
spelling space, invariant space, structural mutants, and character-class portability have each been
swept, by different reviewers, with the same answer.
