package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/snapcache"
	"github.com/relux-works/curator/internal/transaction"
)

const cacheUsage = `Usage:
  curator cache prune [--keep-last <n>] [--older-than <duration>] [--dry-run] [--json]

Remove commit-keyed source snapshots (cache/<source>/<commit>) that nothing
references (manager profile section 10.1). A snapshot is never removed while
an install marker, a Skillfile.lock.json, or an installed profile lock names
its commit, or while it was used within the 24h grace period.

  --keep-last <n>          keep the n most recently used snapshots of every source
  --older-than <duration>  keep snapshots used within the duration (72h, 14d)
  --dry-run                print the plan and remove nothing
  --json                   print the retention report as JSON

With neither policy flag, every unreferenced snapshot outside the grace period
is removed. A dry run refuses pending transaction recovery and exits 1 without
removing anything. If a marker or lock cannot be read, nothing is removed and the
command exits 1. Sizes are allocated bytes; on APFS, clones can make them
overstate what a removal frees.
`

func (c cli) cmdCache(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		_, _ = fmt.Fprint(c.stdout, cacheUsage)
		if len(args) == 0 {
			return exitUsage
		}
		return exitOK
	}
	if args[0] != "prune" {
		_, _ = fmt.Fprintf(c.stderr, "curator: unknown cache subcommand %q\n\n%s", args[0], cacheUsage)
		return exitUsage
	}
	return c.cmdCachePrune(args[1:])
}

func (c cli) cmdCachePrune(args []string) int {
	flags := c.newFlagSet("cache prune")
	keepLast := flags.String("keep-last", "", "keep the n most recently used snapshots of every source")
	olderThan := flags.String("older-than", "", "keep snapshots used within the duration")
	dryRun := flags.Bool("dry-run", false, "print the plan and remove nothing")
	jsonOutput := flags.Bool("json", false, "print the retention report as JSON")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	if len(positional) != 0 {
		_, _ = fmt.Fprintf(c.stderr, "curator: cache prune takes no operands\n\n%s", cacheUsage)
		return exitUsage
	}
	policy := snapcache.Policy{Grace: snapcache.DefaultGrace}
	if *keepLast != "" {
		n, parseErr := strconv.Atoi(*keepLast)
		if parseErr != nil || n < 0 {
			_, _ = fmt.Fprintf(c.stderr, "curator: --keep-last needs a non-negative integer, got %q\n", *keepLast)
			return exitUsage
		}
		policy.KeepLast = &n
	}
	if *olderThan != "" {
		duration, parseErr := parseRetentionDuration(*olderThan)
		if parseErr != nil {
			_, _ = fmt.Fprintf(c.stderr, "curator: --older-than: %v\n", parseErr)
			return exitUsage
		}
		policy.OlderThan = &duration
	}

	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	home := cfg.Home()
	manager, err := managerlock.New(home)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: acquire the manager-home lock:", err)
		return exitFail
	}
	report, err := pruneUnderLock(home, lock, policy, *dryRun)
	if closeErr := lock.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("release the manager-home lock: %w", closeErr))
	}
	if *jsonOutput {
		encoder := json.NewEncoder(c.stdout)
		encoder.SetIndent("", "  ")
		if encodeErr := encoder.Encode(report); encodeErr != nil {
			err = errors.Join(err, encodeErr)
		}
	} else {
		c.printPruneReport(report)
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	if !report.Certain() {
		_, _ = fmt.Fprintln(c.stderr, "curator: cache prune removed nothing: the reference set could not be proven complete; repair or remove the records named above")
		return exitFail
	}
	return exitOK
}

// pruneUnderLock recovers before a real run. A dry run inspects journal names
// and refuses pending recovery, so it cannot delete through recovery or cleanup.
func pruneUnderLock(home string, lock *managerlock.HomeLock, policy snapcache.Policy, dryRun bool) (snapcache.Report, error) {
	engine, err := transaction.New(home)
	if err != nil {
		return snapcache.Report{}, fmt.Errorf("open the install transaction journal: %w", err)
	}
	if dryRun {
		pending, err := engine.HasPendingRecovery(lock)
		if err != nil {
			return snapcache.Report{DryRun: true}, fmt.Errorf("inspect install transaction journals: %w", err)
		}
		if pending {
			return snapcache.Report{DryRun: true}, errors.New("cache prune dry-run requires completed transaction recovery; run cache prune without --dry-run to recover first")
		}
	} else if err := engine.Recover(lock); err != nil {
		return snapcache.Report{}, fmt.Errorf("recover incomplete install transactions: %w", err)
	}
	refs := scopes.CollectSnapshotReferences(home)
	return snapcache.Prune(snapcache.Request{
		Home:       home,
		Lock:       lock,
		Policy:     policy,
		References: snapcache.References{Commits: refs.Commits, Uncertain: refs.Uncertain},
		Now:        time.Now(),
		DryRun:     dryRun,
	})
}

func (c cli) printPruneReport(report snapcache.Report) {
	for _, warning := range report.Warnings {
		_, _ = fmt.Fprintln(c.stderr, "curator: warning:", warning)
	}
	verb := "remove"
	if report.DryRun {
		verb = "would remove"
	}
	for _, entry := range report.Entries {
		action := "keep"
		if entry.Action == snapcache.ActionRemove {
			action = verb
			if !report.DryRun && !entry.Removed {
				action = "failed to remove"
			}
		}
		_, _ = fmt.Fprintf(c.stdout, "%-14s %s@%s  %9s  last used %s  (%s)\n",
			action, entry.Source, shortCommit(entry.Commit), formatBytes(entry.AllocatedBytes),
			entry.LastUsedAt, strings.ReplaceAll(entry.Reason, "_", " "))
	}
	summary := "removed"
	if report.DryRun {
		summary = "would remove"
	}
	_, _ = fmt.Fprintf(c.stdout, "cache prune: %s %d of %d snapshot%s, %s of %s allocated\n",
		summary, report.Totals.RemoveEntries, report.Totals.Entries, pluralS(report.Totals.Entries),
		formatBytes(report.Totals.RemoveAllocatedBytes), formatBytes(report.Totals.AllocatedBytes))
}

// parseRetentionDuration accepts a Go duration or a whole number of days
// ("14d").
func parseRetentionDuration(value string) (time.Duration, error) {
	if days, ok := strings.CutSuffix(value, "d"); ok {
		n, err := strconv.Atoi(days)
		if err != nil || n < 0 || n > 100000 {
			return 0, fmt.Errorf("%q is not a whole number of days", value)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration < 0 {
		return 0, fmt.Errorf("%q is not a non-negative duration such as 72h or 14d", value)
	}
	return duration, nil
}

func shortCommit(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, suffix := float64(n), "KMGTPE"
	index := -1
	for value >= unit && index < len(suffix)-1 {
		value /= unit
		index++
	}
	return fmt.Sprintf("%.1f %ciB", value, suffix[index])
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
