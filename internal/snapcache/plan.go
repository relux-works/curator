// Package snapcache applies manager profile section 10.1 snapshot-cache
// retention to the commit-keyed snapshot store at
// cache/<source>/<commit>/snapshot: inventory with sizes and last-use times,
// the retention decision, and move-then-delete removal.
package snapcache

import (
	"bytes"
	"sort"
	"time"
)

// Reasons, in the section 10.1 decision order. Each entry carries exactly
// one, from the first rule that applies.
const (
	ReasonReachable          = "reachable"
	ReasonReferenceUncertain = "reference_uncertain"
	ReasonGrace              = "grace"
	ReasonKeepLast           = "keep_last"
	ReasonNewerThan          = "newer_than"
	ReasonUnreachable        = "unreachable"
)

// Actions.
const (
	ActionRetain = "retain"
	ActionRemove = "remove"
)

// DefaultGrace is Curator's documented grace period: an entry used more
// recently than this is retained, because a concurrent resolution may have
// published it before recording any reference. It matches the build-cache
// sweep's grace.
const DefaultGrace = 24 * time.Hour

// Policy is one invocation's operator retention policy.
type Policy struct {
	// KeepLast, when set, retains that many most recently used entries of
	// every source. Nil or zero adds no keep-last retention.
	KeepLast *int
	// OlderThan, when set, retains every entry whose age is at most its value.
	OlderThan *time.Duration
	// Grace retains every entry younger than it.
	Grace time.Duration
}

// Entry is one commit-keyed snapshot.
type Entry struct {
	Source   string
	Commit   string
	LastUsed time.Time
	// Reachable reports whether a live record names the commit.
	Reachable      bool
	AllocatedBytes int64
	LogicalBytes   int64
	// Path is the entry directory (cache/<source>/<commit>); empty in pure
	// decision tests.
	Path string
}

// Decision is one entry with its planned action and reason.
type Decision struct {
	Entry
	Action string
	Reason string
}

// Plan applies the section 10.1 decision rule. certain reports whether the
// reference set is provably complete. The result is ordered by source, then
// commit, by bytes.
func Plan(entries []Entry, certain bool, policy Policy, now time.Time) []Decision {
	keep := 0
	if policy.KeepLast != nil {
		keep = *policy.KeepLast
	}
	window := keepLastWindow(entries, keep)
	decisions := make([]Decision, 0, len(entries))
	for _, entry := range entries {
		age := now.Sub(entry.LastUsed)
		var reason string
		switch {
		case entry.Reachable:
			reason = ReasonReachable
		case !certain:
			reason = ReasonReferenceUncertain
		case age < policy.Grace:
			reason = ReasonGrace
		case window[entryKey{entry.Source, entry.Commit}]:
			reason = ReasonKeepLast
		case policy.OlderThan != nil && age <= *policy.OlderThan:
			reason = ReasonNewerThan
		default:
			reason = ReasonUnreachable
		}
		action := ActionRetain
		if reason == ReasonUnreachable {
			action = ActionRemove
		}
		decisions = append(decisions, Decision{Entry: entry, Action: action, Reason: reason})
	}
	sort.Slice(decisions, func(i, j int) bool {
		if c := bytes.Compare([]byte(decisions[i].Source), []byte(decisions[j].Source)); c != 0 {
			return c < 0
		}
		return bytes.Compare([]byte(decisions[i].Commit), []byte(decisions[j].Commit)) < 0
	})
	return decisions
}

type entryKey struct{ source, commit string }

// keepLastWindow selects, per source, the keep entries with the latest
// last-use time, ties broken by commit ascending by bytes. Reachable entries
// count toward the window.
func keepLastWindow(entries []Entry, keep int) map[entryKey]bool {
	window := map[entryKey]bool{}
	if keep <= 0 {
		return window
	}
	bySource := map[string][]Entry{}
	for _, entry := range entries {
		bySource[entry.Source] = append(bySource[entry.Source], entry)
	}
	for source, group := range bySource {
		sort.Slice(group, func(i, j int) bool {
			if !group[i].LastUsed.Equal(group[j].LastUsed) {
				return group[i].LastUsed.After(group[j].LastUsed)
			}
			return bytes.Compare([]byte(group[i].Commit), []byte(group[j].Commit)) < 0
		})
		for index := 0; index < keep && index < len(group); index++ {
			window[entryKey{source, group[index].Commit}] = true
		}
	}
	return window
}
