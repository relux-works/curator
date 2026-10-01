package snapcache

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HomeLock witnesses the caller-held manager-home mutation lock.
type HomeLock interface {
	AssertHeld() error
}

// References is the section 10.1 reference set.
type References struct {
	Commits   map[string]bool
	Uncertain []string
}

// Request is one retention invocation.
type Request struct {
	Home       string
	Lock       HomeLock
	Policy     Policy
	References References
	Now        time.Time
	DryRun     bool
}

// ReportPolicy is the policy member of the retention report.
type ReportPolicy struct {
	KeepLast         *int   `json:"keep_last"`
	OlderThanSeconds *int64 `json:"older_than_seconds"`
	GraceSeconds     int64  `json:"grace_seconds"`
}

// ReportEntry is one entry of the retention report.
type ReportEntry struct {
	Source         string `json:"source"`
	Commit         string `json:"commit"`
	LastUsedAt     string `json:"last_used_at"`
	AllocatedBytes int64  `json:"allocated_bytes"`
	LogicalBytes   int64  `json:"logical_bytes"`
	Reachable      bool   `json:"reachable"`
	Action         string `json:"action"`
	Reason         string `json:"reason"`
	Removed        bool   `json:"removed"`
}

// ReportTotals sums the report.
type ReportTotals struct {
	Entries              int   `json:"entries"`
	AllocatedBytes       int64 `json:"allocated_bytes"`
	LogicalBytes         int64 `json:"logical_bytes"`
	RemoveEntries        int   `json:"remove_entries"`
	RemoveAllocatedBytes int64 `json:"remove_allocated_bytes"`
	RemoveLogicalBytes   int64 `json:"remove_logical_bytes"`
}

// Report is the section 10.1 retention report.
type Report struct {
	referenceSetCertain bool
	DryRun              bool          `json:"dry_run"`
	Policy              ReportPolicy  `json:"policy"`
	SizeMeasure         string        `json:"size_measure"`
	Entries             []ReportEntry `json:"entries"`
	Totals              ReportTotals  `json:"totals"`
	Warnings            []string      `json:"warnings"`
}

// Certain reports whether the plan was computed over a provably complete
// reference set.
func (report Report) Certain() bool {
	return report.referenceSetCertain
}

// Prune computes the section 10.1 plan and, unless DryRun, removes every
// entry it marks for removal. The caller holds the manager-home mutation lock
// and has already recovered incomplete transactions.
//
// A removal first renames the entry out of the store namespace, then deletes
// it, so an interrupted removal never leaves a partial tree at the entry's
// canonical path. The report is complete even when a removal fails; the
// error names every failure.
func Prune(request Request) (Report, error) {
	report := Report{
		referenceSetCertain: len(request.References.Uncertain) == 0,
		DryRun:              request.DryRun,
		SizeMeasure:         "allocated",
		Entries:             []ReportEntry{},
		Warnings:            []string{},
		Policy: ReportPolicy{
			KeepLast:     request.Policy.KeepLast,
			GraceSeconds: int64(request.Policy.Grace / time.Second),
		},
	}
	if olderThan := request.Policy.OlderThan; olderThan != nil {
		seconds := int64(*olderThan / time.Second)
		report.Policy.OlderThanSeconds = &seconds
	}
	if request.Home == "" {
		return report, errors.New("manager home is empty")
	}
	if request.Lock == nil {
		return report, errors.New("caller-held manager-home lock is required")
	}
	if err := request.Lock.AssertHeld(); err != nil {
		return report, fmt.Errorf("manager-home lock is not held: %w", err)
	}
	inventory, err := ReadInventory(request.Home)
	if err != nil {
		return report, fmt.Errorf("read the snapshot cache: %w", err)
	}
	report.Warnings = append(report.Warnings, request.References.Uncertain...)
	report.Warnings = append(report.Warnings, inventory.Notes...)
	for index := range inventory.Entries {
		inventory.Entries[index].Reachable = request.References.Commits[inventory.Entries[index].Commit]
	}
	certain := report.Certain()
	decisions := Plan(inventory.Entries, certain, request.Policy, request.Now)

	var failures []error
	for _, leftover := range inventory.Leftovers {
		if !certain {
			report.Warnings = append(report.Warnings, fmt.Sprintf("interrupted removal %s retained because references are uncertain", leftover))
			continue
		}
		report.Warnings = append(report.Warnings, fmt.Sprintf("interrupted removal %s scheduled for cleanup", leftover))
		if request.DryRun {
			continue
		}
		if err := os.RemoveAll(leftover); err != nil {
			failures = append(failures, fmt.Errorf("delete interrupted removal %s: %w", leftover, err))
			continue
		}
	}
	for _, decision := range decisions {
		row := ReportEntry{
			Source:         decision.Source,
			Commit:         decision.Commit,
			LastUsedAt:     decision.LastUsed.UTC().Format(time.RFC3339),
			AllocatedBytes: decision.AllocatedBytes,
			LogicalBytes:   decision.LogicalBytes,
			Reachable:      decision.Reachable,
			Action:         decision.Action,
			Reason:         decision.Reason,
		}
		report.Totals.Entries++
		report.Totals.AllocatedBytes += row.AllocatedBytes
		report.Totals.LogicalBytes += row.LogicalBytes
		if row.Action == ActionRemove {
			report.Totals.RemoveEntries++
			report.Totals.RemoveAllocatedBytes += row.AllocatedBytes
			report.Totals.RemoveLogicalBytes += row.LogicalBytes
			if !request.DryRun {
				if err := removeEntry(CacheRoot(request.Home), decision.Path); err != nil {
					failures = append(failures, fmt.Errorf("remove snapshot %s@%s: %w", decision.Source, decision.Commit, err))
				} else {
					row.Removed = true
				}
			}
		}
		report.Entries = append(report.Entries, row)
	}
	return report, errors.Join(failures...)
}

// removeEntry moves entry out of the store namespace, deletes it, and then
// removes source directories the removal left empty, up to the cache root.
func removeEntry(cacheRoot, entry string) error {
	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return err
	}
	parent := filepath.Dir(entry)
	moved := filepath.Join(parent, prunePrefix+filepath.Base(entry)+"-"+hex.EncodeToString(suffix[:]))
	if err := os.Rename(entry, moved); err != nil {
		return fmt.Errorf("move out of the store: %w", err)
	}
	if err := os.RemoveAll(moved); err != nil {
		return fmt.Errorf("delete %s (a later run finishes it): %w", moved, err)
	}
	for dir := parent; dir != cacheRoot && isWithin(cacheRoot, dir); dir = filepath.Dir(dir) {
		if err := os.Remove(dir); err != nil {
			break // not empty, or already gone
		}
	}
	return nil
}

func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && rel != ".." && !filepath.IsAbs(rel) &&
		!strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
