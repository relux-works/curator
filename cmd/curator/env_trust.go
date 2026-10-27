package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/operatorpolicy"
	"github.com/relux-works/curator/internal/projectcontext"
)

const envTrustUsage = `usage: curator env trust <key|grant|revoke|list|read> [flags]
  key --principal <id> --public-key <base64-ed25519>  one-time operator enrollment
  grant --command <signed-command.json>             trust and digest-listed hooks
  revoke --command <signed-command.json>            revoke every event for an entry
  list                                             inspect private policy/event history
  read <env> <absolute-project-root> --checkout <id> --lane <id> --profile <name> --lock-sha256 <hex>
Signed commands require operator and actor signatures; package declarations cannot authorize them.
No vendor adapter is currently qualified for unattended policy admission.
`

func (c cli) cmdEnvTrust(cfg *config.Config, args []string) int {
	if len(args) == 0 || wantsHelp(args) {
		_, _ = fmt.Fprint(c.stdout, envTrustUsage)
		return exitOK
	}
	store := operatorpolicy.Store{Home: cfg.Home()}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var output any
	var err error
	flags := c.newFlagSet("env trust " + args[0])
	switch args[0] {
	case "key":
		principal := flags.String("principal", "", "operator principal")
		key := flags.String("public-key", "", "base64 Ed25519 public key")
		rest, e := parseInterspersed(flags, args[1:])
		if e != nil || len(rest) != 0 || *principal == "" || *key == "" {
			return c.envTrustUsageError()
		}
		err = store.PinOperator(ctx, *principal, *key)
		output = map[string]string{"status": "operator-key-enrolled"}
	case "grant", "revoke":
		command := flags.String("command", "", "signed operator command file")
		rest, e := parseInterspersed(flags, args[1:])
		if e != nil || len(rest) != 0 || *command == "" {
			return c.envTrustUsageError()
		}
		signed, e := operatorpolicy.ReadSignedCommand(*command)
		if e != nil || signed.Command.Operation != args[0] {
			_, _ = fmt.Fprintln(c.stderr, "workspace_trust_policy_authority: signed command unreadable or operation mismatch")
			return exitFail
		}
		output, err = store.Apply(ctx, signed)
	case "list":
		if len(args) != 1 {
			return c.envTrustUsageError()
		}
		output, err = store.List(ctx)
	case "read":
		checkout := flags.String("checkout", "", "protected checkout ID")
		lane := flags.String("lane", "", "exact lane ID")
		profile := flags.String("profile", "", "selected profile")
		lock := flags.String("lock-sha256", "", "current lock digest")
		rest, e := parseInterspersed(flags, args[1:])
		if e != nil || len(rest) != 2 || *checkout == "" || *lane == "" || *profile == "" || *lock == "" {
			return c.envTrustUsageError()
		}
		output, err = store.Read(ctx, operatorpolicy.Selection{Checkout: *checkout, Path: rest[1], Environment: rest[0], Lane: *lane, Profile: projectcontext.Profile{Name: *profile, LockSHA256: *lock}})
	default:
		return c.envTrustUsageError()
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, err)
		return exitFail
	}
	if err = json.NewEncoder(c.stdout).Encode(output); err != nil {
		_, _ = fmt.Fprintln(c.stderr, err)
		return exitFail
	}
	return exitOK
}
func (c cli) envTrustUsageError() int { _, _ = fmt.Fprint(c.stderr, envTrustUsage); return exitUsage }
