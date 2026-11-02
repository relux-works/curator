package homelease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/homelease/state"
)

type Result struct {
	SchemaVersion int            `json:"schema_version"`
	Operation     string         `json:"operation"`
	OK            bool           `json:"ok"`
	Exit          int            `json:"exit"`
	Code          *string        `json:"code"`
	Reason        string         `json:"reason"`
	Result        any            `json:"result"`
	Details       *state.Details `json:"details"`
}

// RunCLI owns the closed recognized lease-family envelope, including usage and
// setup failures. Unknown routing preserves the ordinary human CLI refusal.
func RunCLI(ctx context.Context, args []string, setup func() (*Manager, error), stdout, stderr io.Writer) int {
	if len(args) == 0 || !knownOperation(args[0]) {
		_, _ = fmt.Fprintln(stderr, "curator: env lease requires acquire, bind, release, list or break")
		return 2
	}
	op := args[0]
	r := Result{SchemaVersion: 1, Operation: op}
	finish := func(err error) int {
		if err != nil {
			var refusal *Refusal
			if !errors.As(err, &refusal) {
				refusal = state.Refuse("environment_lease_authority_refused", "authority_unavailable")
			}
			r.Exit = 1
			if refusal.Code == "usage" {
				r.Exit = 2
			}
			r.Code = &refusal.Code
			r.Reason = refusal.Reason
			r.Details = &refusal.Details
			r.Result = nil
			_, _ = fmt.Fprintln(stderr, "curator:", refusal)
		} else {
			r.OK = true
		}
		raw, encodeErr := state.Canonical(r)
		if encodeErr != nil || state.Validate("result", raw) != nil {
			_, _ = fmt.Fprintln(stderr, "curator: home lease result could not be encoded")
			return 1
		}
		if _, err := fmt.Fprintln(stdout, string(raw)); err != nil {
			return 1
		}
		return r.Exit
	}
	flags, err := parseFlags(op, args[1:])
	if err != nil {
		return finish(err)
	}
	var holder Holder
	if op == "acquire" || op == "bind" {
		holder, err = state.ParseHolder([]byte(flags["holder"]))
		if err != nil {
			return finish(err)
		}
	}
	var revision int64
	if op == "bind" || op == "break" {
		revision, err = strconv.ParseInt(flags["if-revision"], 10, 64)
		if err != nil || revision < 1 {
			return finish(state.Refuse("usage", "revision_required"))
		}
	}
	m, err := setup()
	if err != nil {
		return finish(err)
	}
	var out Outcome
	switch op {
	case "acquire":
		out, r.Reason, err = m.Acquire(ctx, Request{Environment: envregistry.NormalizeEnvID(flags["env"]), Profile: flags["profile"], LockSHA256: flags["lock-sha256"], RequestID: flags["request-id"], Holder: holder})
		r.Result = out
	case "bind":
		out, r.Reason, err = m.Bind(ctx, flags["lease"], revision, holder)
		r.Result = out
	case "release":
		out, r.Reason, err = m.Release(ctx, flags["lease"])
		r.Result = out
	case "break":
		out, r.Reason, err = m.Break(ctx, flags["lease"], revision, flags["operator"] == "true", flags["reason"])
		r.Result = out
	case "list":
		var rows []Record
		rows, err = m.List(ctx, envregistry.NormalizeEnvID(flags["env"]), flags["profile"])
		r.Reason = "listed"
		r.Result = struct {
			Leases []Record `json:"leases"`
		}{rows}
	}
	return finish(err)
}

func knownOperation(op string) bool {
	switch op {
	case "acquire", "bind", "release", "list", "break":
		return true
	}
	return false
}
func parseFlags(op string, args []string) (map[string]string, error) {
	allowed := map[string]bool{"format": true}
	required := []string{}
	switch op {
	case "acquire":
		required = []string{"env", "profile", "lock-sha256", "request-id", "holder"}
	case "bind":
		required = []string{"lease", "if-revision", "holder"}
	case "release":
		required = []string{"lease"}
	case "break":
		required = []string{"lease", "if-revision", "operator", "reason"}
	case "list":
		allowed["env"] = true
		allowed["profile"] = true
	}
	for _, key := range required {
		allowed[key] = true
	}
	flags := map[string]string{}
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			return nil, state.Refuse("usage", "invalid_flags")
		}
		key, value, hasValue := strings.Cut(strings.TrimPrefix(args[i], "--"), "=")
		if !allowed[key] {
			return nil, state.Refuse("usage", "invalid_flags")
		}
		if _, exists := flags[key]; exists {
			return nil, state.Refuse("usage", "invalid_flags")
		}
		if key == "operator" {
			if hasValue {
				return nil, state.Refuse("usage", "invalid_flags")
			}
			value = "true"
		} else if !hasValue {
			i++
			if i == len(args) || strings.HasPrefix(args[i], "--") {
				return nil, state.Refuse("usage", "invalid_flags")
			}
			value = args[i]
		}
		flags[key] = value
	}
	if format, ok := flags["format"]; ok && format != "json" {
		return nil, state.Refuse("usage", "invalid_format")
	}
	for _, key := range required {
		if flags[key] != "" {
			continue
		}
		reason := "invalid_selector"
		switch key {
		case "operator":
			reason = "operator_confirmation_required"
		case "reason":
			reason = "reason_required"
		case "if-revision":
			reason = "revision_required"
		case "holder":
			reason = "invalid_holder"
		}
		return nil, state.Refuse("usage", reason)
	}
	return flags, nil
}

// DecodeResult is a closed verifier for command consumers; successful terminal
// retrieval remains non-admitting even though its transport exit is zero.
func DecodeResult(raw []byte) (Result, error) {
	var r Result
	if err := state.Validate("result", raw); err != nil {
		return r, err
	}
	err := json.Unmarshal(raw, &r)
	return r, err
}
