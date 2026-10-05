# Atomic profile identity migration (rc.4)

The rc.3 release remains a landing prerequisite; integration owns tag ordering.
This change leaves release pinning, seed/posture revisions, and release notes alone.

WriteVersion reaches install target planning/publication (install/targets.go,
install.go), core marker shape and content identity (marker.go), context resolver
state pins (contextresolve.go), state-store acquisition (contextstore.go), default
lock shape (contextlock.go), materialized surface hashing (contextmaterialize.go),
and managed-home marker/credential publication (managed.go, migrate.go).
New installs use v2. Existing carriers select their own reader version; a writer
switch never changes the interpretation of a stored pin.

The atomic unit is a profile lock plus every provisioned managed environment for that
profile: immutable state trees are verified with the old framing, copied and
rehash-keyed with v2, then the lock, generated documents/headers, surface entries,
and schema-3 markers publish in one durable transaction under the home lock.
Machine-native in-place projections retain their explicit old version and become
stale when the lock changes, as on an ordinary profile update; Use regenerates
them with the lock-declared framing. Native credential and seed files stay
tool-owned. Old store entries remain available to other profiles and the previous lock.
New immutable cache entries may survive a refused publication; they do not
activate a profile. Credential migration includes its requested link changes and
credential records in this same transaction when a legacy profile is involved.
Resolve --repair uses the same profile migration before ordinary repair.
Use performs the same identity migration before its existing native scope
materialization; partial native switch semantics remain those of M11.

The transaction engine owns exact entry preimages, rollback, and restart recovery:
a syscall/commit failure restores the old lock/surfaces/markers/credential links;
an interrupted durable commit resumes under the next home mutation lock. No
seed files or credential payloads are read/copied. Lock-free readers may observe
a transient mismatch and must refuse it, never accept an inconsistent profile.
Migration plans bind the lock, immutable store and all affected managed surface
entry identities, including siblings outside the credential environment filter.

State pins change only after verified snapshot rehashing. Git commit pins remain
commits. Generated headers bind the new canonical lock hash. Every surface is
regenerated from the pinned store and its v2 hash recomputed. Prior seed and
credential metadata survive unless the explicit credential operation changes
it. A marker/lock version mismatch is refused; neither v1 hashes labelled v2 nor
v2 hashes labelled v1 are fallback identities.

Validation is limited to targeted GOFLAGS=-work tests and a CLI build. Full
suite, cross-platform/race and hosted lint and hosted acceptance are the Change Request
gate's responsibility. Evidence records actual exit codes and does not assert
hosted green before that gate runs.
