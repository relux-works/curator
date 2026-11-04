package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/relux-works/curator/internal/config"
)

const claudeAuthConfigUsage = `Metadata-only Claude OAuth configuration:
  curator config claude-auth plan (--keychain-service S --keychain-account A | --file PATH) --out CANDIDATE
  curator config claude-auth apply <same store flags> --expect-source-sha256 X --expect-result-sha256 Y --expect-policy-sha256 P
  curator config claude-auth restore --backup PATH --expect-current-sha256 Y

Plan writes a new private candidate; apply retains a private exact-byte backup.
Apply and restore require exclusive custody: final recheck plus rename is not
atomic compare-and-swap against noncooperating writers. Credentials are not read.
`

func (c cli) cmdConfigClaudeAuth(args []string) int {
	usageError := func() int { _, _ = fmt.Fprint(c.stderr, claudeAuthConfigUsage); return exitUsage }
	if len(args) == 0 {
		return usageError()
	}
	allowed := map[string]bool{}
	switch args[0] {
	case "plan":
		for _, key := range []string{"file", "keychain-service", "keychain-account", "out"} {
			allowed[key] = true
		}
	case "apply":
		for _, key := range []string{"file", "keychain-service", "keychain-account", "expect-source-sha256", "expect-result-sha256", "expect-policy-sha256"} {
			allowed[key] = true
		}
	case "restore":
		for _, key := range []string{"backup", "expect-current-sha256"} {
			allowed[key] = true
		}
	default:
		return usageError()
	}
	// Validate flag names and repetition before flag.Parse, whose ordinary
	// diagnostic would reflect rejected values. Every admitted flag has a value.
	seen := map[string]bool{}
	for i := 1; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			return usageError()
		}
		name, _, assigned := strings.Cut(strings.TrimPrefix(args[i], "--"), "=")
		if !allowed[name] || seen[name] {
			return usageError()
		}
		seen[name] = true
		if !assigned {
			i++
			if i >= len(args) {
				return usageError()
			}
		}
	}
	fs := flag.NewFlagSet("claude-auth", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	values := map[string]*string{}
	for name := range allowed {
		values[name] = fs.String(name, "", "")
	}
	if fs.Parse(args[1:]) != nil || fs.NArg() != 0 {
		return usageError()
	}
	value := func(name string) string {
		if values[name] == nil {
			return ""
		}
		return *values[name]
	}
	for name := range seen {
		if value(name) == "" {
			return usageError()
		}
	}
	selection := config.LauncherClaudeAuth{Type: "oauth"}
	if args[0] != "restore" {
		if seen["file"] {
			if seen["keychain-service"] || seen["keychain-account"] {
				return usageError()
			}
			selection.Store = config.LauncherAuthStore{Type: "file", Path: value("file")}
		} else {
			if !seen["keychain-service"] || !seen["keychain-account"] {
				return usageError()
			}
			selection.Store = config.LauncherAuthStore{Type: "keychain", Service: value("keychain-service"), Account: value("keychain-account")}
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	printJSON := func(v any) bool {
		raw, err := json.Marshal(v)
		if err != nil {
			return false
		}
		_, err = fmt.Fprintln(c.stdout, string(raw))
		return err == nil
	}
	var result *config.ClaudeAuthEditResult
	var err error
	switch args[0] {
	case "plan":
		if !seen["out"] {
			return usageError()
		}
		var plan *config.ClaudeAuthEditPlan
		plan, err = config.PlanClaudeAuth(c.config.Path(), selection)
		if err == nil {
			err = config.WriteClaudeAuthCandidate(plan, value("out"))
		}
		if err == nil {
			origin := "selected-source"
			if c.config.Path() == config.UserPath() {
				origin = "default"
				if os.Getenv("CURATOR_CONFIG") != "" {
					origin = "CURATOR_CONFIG"
				}
			}
			if printJSON(struct {
				*config.ClaudeAuthEditPlan
				SelectionSource string `json:"selection_source"`
			}{plan, origin}) {
				return exitOK
			}
			_, _ = fmt.Fprintln(c.stderr, "curator: candidate created; plan reporting failed")
			return exitFail
		}
	case "apply":
		if !seen["expect-source-sha256"] || !seen["expect-result-sha256"] || !seen["expect-policy-sha256"] {
			return usageError()
		}
		result, err = config.ApplyClaudeAuth(ctx, config.ClaudeAuthEditRequest{Path: c.config.Path(), Selection: selection, ExpectSourceSHA256: value("expect-source-sha256"), ExpectResultSHA256: value("expect-result-sha256"), ExpectPolicySHA256: value("expect-policy-sha256")})
	case "restore":
		if !seen["backup"] || !seen["expect-current-sha256"] {
			return usageError()
		}
		result, err = config.RestoreClaudeAuth(ctx, config.ClaudeAuthRestoreRequest{Path: c.config.Path(), BackupPath: value("backup"), ExpectCurrentSHA256: value("expect-current-sha256")})
	}
	if result != nil && !printJSON(result) {
		// A broken output stream cannot turn a completed rename into refusal.
		_, _ = fmt.Fprintln(c.stderr, "curator:", result.Status, "(result reporting failed)")
		return exitFail
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
