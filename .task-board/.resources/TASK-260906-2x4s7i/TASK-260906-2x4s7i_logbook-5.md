# Logbook — cycle-5 review of PR #47 (`407424e`)

Kept out of the repository: this was a read-only review and the branch must gain no stray tracked
file.

## 1. Identical blobs are the cheapest proof that a delta changed no behaviour

The cycle-4 delta touched seven files, five of them generator output. Rather than re-driving the
whole classification matrix at both heads, `git rev-parse <commit>:<path>` on the schema,
`protocol/environments.md` and `profiles/manager.md` showed all three byte-identical between
`18dca85` and `407424e`. Combined with a purely additive corpus diff (no existing `valid` flag
touched), that settles "no other case's verdict moved" in three commands instead of two full sweeps.
The committed schema's SHA-256 also matched the digest cycle 4 recorded independently — a free
cross-cycle consistency check.

## 2. `git hash-object` applies the *outer* repository's filters

Verifying an `rsync` copy against `git ls-tree` reported two mismatches on
`conformance/v1/fixtures/byte-exact/{crlf,mixed}.txt`. The bytes were identical under `xxd`. The cause
is that the scratch copy lived inside another worktree, so `git hash-object` resolved `.gitattributes`
from the enclosing repository and applied `eol=lf` to files whose whole purpose is committed CRLF.
`git hash-object --no-filters` is the only correct comparison for a byte-exactness fixture. Worth
remembering alongside the existing note that `git archive` breaks the same fixture.

## 3. Case-narrowing is its own mutant family, and it found the last real gap

Four cycles produced deletion, widening, unanchoring and character-class mutants of the discriminator.
None of them narrowed a class *by case*. `^[A-Za-z]:[\\/]` → `^[A-Z]:[\\/]` survived the whole corpus
and flips 64 spellings from `path` to `refused`, the class being `c:\users\operator\context` — an
ordinary Windows spelling, and precisely the F1 class the whole change exists to remove. Every
published Windows positive spells the drive `C`. Add case-narrowing to the standard mutant vocabulary
for any pattern containing a case-insensitive class.

## 4. "Every claim resolves against a published case" catches enumerations, not parentheticals

The changelog's enumerated list of pinned spellings checked out twelve for twelve. The two inaccurate
statements were both outside it: a blanket sentence ("a `://` URL outside the four schemes is refused
outright" — false for a one-character scheme) and a parenthetical ("and no backslash" — the encoded
rule is *no backslash immediately after the SCP colon*). The same paragraph elsewhere states the
precise version of the second rule. A prose gate has to read the connective sentences, not only the
lists that look checkable.

## 5. core §6.1's mandate is "not treated as local", not "rejected"

The schema refuses `git@my_host:x` and `github.com:\example\x` outright but admits `https://h:22/x`,
`https://user:pw@h/x`, `https://h/x?q=1` and `https://h/x\y` as `git`-with-a-form. That asymmetry
looks arbitrary until the sentence is named: §6.1 says "Invalid network forms MUST be rejected, **not
treated as local**". Refusal is forced only where the alternative classification would be `path`. A
`://` form is already `git`, so the mandate is satisfied and §1.1 reports `profile_source_invalid` at
resolution. Reviewing the rule instead of the shape avoided a false finding.

## 6. On Windows, `C:` is drive-relative, not absolute

The reasoning that decided the new negative case only works if you know that `C:` denotes the current
directory *of drive C* — per-process state — rather than the drive root. §1 admits an absolute or a
project-relative path; a drive-relative reference is neither. The POSIX reading, where `C:` is a legal
relative filename, is the residual bound, and it is the same one the change already states as "cannot
distinguish a directory spelled like a host from a real host".
