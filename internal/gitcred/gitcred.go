// Package gitcred reads and writes operator credentials through the
// operator's own Git credential machinery.
//
// Every call is `git credential fill|approve|reject`. That is the one
// mechanism which exists identically on macOS, Windows and Linux, it speaks
// to whichever helper the operator already configured — osxkeychain, wincred,
// libsecret, or a credential manager — and it adds no dependency to the
// manager. The helper is selected by the operator's own Git configuration,
// never by a manifest, descriptor, repository, substitution or marker, which
// is what keeps this material operator-owned (Spec §12.2).
//
// Two entries per host are addressed here and they never collide:
//
//   - the operator's own HTTPS credential for a host, read as it stands;
//   - a manager-namespaced entry, stored under a username carrying
//     NamespacePrefix so the manager's own material is a separate record from
//     the operator's, on every platform's helper.
//
// Interactive prompting is disabled on every call: an absent credential
// degrades into "nothing here" rather than blocking a run on a dialog nobody
// is watching. A read that finds nothing is therefore not an error. A write
// is: it is always proved by reading the value back, because a helper can
// report a successful write while persisting nothing, and an unverified write
// would look like a configured scope that fails later, mid-install.
//
// The manager runs these calls itself, outside the fetch process graph. The
// fetch owns a private HOME, so the operator home is pinned explicitly here;
// without it the helper the operator configured is simply not found.
package gitcred

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// NamespacePrefix opens the username of every manager-stored entry. A
// distinct username is what keeps a manager entry separate from the
// operator's own credential for the same host, so neither overwrites nor
// answers for the other.
const NamespacePrefix = "curator-build-https:"

// DefaultTimeout bounds one credential call. A helper that talks to a locked
// keychain can hang indefinitely; the manager treats that as no material.
const DefaultTimeout = 15 * time.Second

// drainDelay bounds the wait after a call's deadline has killed it.
//
// The deadline kills Git, and a credential helper is not Git: it is the
// child Git started, it inherits Git's output, and it outlives the kill. A
// call left waiting on the pipes that orphan still holds would return only
// when the orphan does — which, for the wedged helper the deadline exists
// for, is never. So the call stops waiting on it, and one credential call
// costs at most a timeout plus this.
const drainDelay = 2 * time.Second

// DefaultUsername is the username a host credential is reported under when
// the operator's helper names none.
const DefaultUsername = "token"

// maxAnswerBytes bounds what one helper answer may occupy.
const maxAnswerBytes = 64 << 10

// NamespaceUsername is the username a manager-stored credential for scope
// lives under.
func NamespaceUsername(scope string) string { return NamespacePrefix + scope }

// ScopeHost is the host part of a canonical-identity scope, which is the host
// its credential is addressed by.
func ScopeHost(scope string) string {
	host, _, _ := strings.Cut(scope, "/")
	return host
}

// Access performs credential calls against one Git tool and one operator
// home. The zero value is usable: it resolves `git` on PATH, pins the
// operator home this process runs under, and inherits this process's
// environment.
type Access struct {
	// Git is the Git executable to call. Empty resolves `git` on PATH.
	Git string
	// Home is the operator home whose Git configuration selects the helper.
	// Empty resolves the home of the account this process runs under.
	Home string
	// Environ is the base environment for the call. Nil inherits this
	// process's environment, which is what carries the session state a
	// helper needs — a desktop bus, a keychain session, a proxy.
	Environ []string
	// Timeout bounds one call. Zero means DefaultTimeout.
	Timeout time.Duration
}

// HostCredential is the operator's own credential for one host.
type HostCredential struct {
	Username string
	Secret   string
}

// HostMaterial is a presence-only view of the operator material one host
// could authenticate with. It names what exists and never retains a secret,
// so a prompt can describe the operator's options without holding any.
type HostMaterial struct {
	// HostCredential reports that the operator's own helper holds a
	// credential for the host.
	HostCredential bool
	// HostUsername is the username that credential is held under. It is not
	// a secret and it is what makes an operator's own entry recognisable in
	// a prompt.
	HostUsername string
	// Scopes are the scopes a manager-stored entry already exists for, in
	// the order they were asked about.
	Scopes []string
}

// Empty reports a host no operator material was found for at all.
func (m HostMaterial) Empty() bool { return !m.HostCredential && len(m.Scopes) == 0 }

// ReadHost reads the operator's own HTTPS credential for host.
//
// A manager-namespaced entry is never reported as the operator's own: a
// helper asked for a host without a username answers with whatever record it
// holds for that host, which can be a manager entry, and reporting that back
// as the operator's own credential would make one record look like two.
func (a Access) ReadHost(ctx context.Context, host string) (HostCredential, bool) {
	if !portableValue(host) {
		return HostCredential{}, false
	}
	answer, ok := a.call(ctx, "fill", field{"protocol", "https"}, field{"host", host})
	if !ok || answer["password"] == "" {
		return HostCredential{}, false
	}
	username := answer["username"]
	if strings.HasPrefix(username, NamespacePrefix) {
		return HostCredential{}, false
	}
	if username == "" {
		username = DefaultUsername
	}
	return HostCredential{Username: username, Secret: answer["password"]}, true
}

// ReadScoped reads the manager-stored credential for scope on host.
//
// The answer is accepted only when the helper answers for the username that
// was asked about. A helper free to answer a near miss must not hand the
// operator's own credential back as the manager's own entry.
func (a Access) ReadScoped(ctx context.Context, scope, host string) (string, bool) {
	if !portableValue(scope) || !portableValue(host) {
		return "", false
	}
	username := NamespaceUsername(scope)
	answer, ok := a.call(ctx, "fill",
		field{"protocol", "https"}, field{"host", host}, field{"username", username})
	if !ok || answer["password"] == "" {
		return "", false
	}
	if answered := answer["username"]; answered != "" && answered != username {
		return "", false
	}
	return answer["password"], true
}

// StoreScoped saves secret as the manager-stored credential for scope on
// host, and proves it was saved by reading it back.
//
// The read-back is not belt and braces. A credential helper reports success
// for a write it never persisted — a manager fronting a platform store it
// cannot reach does exactly that, which is the normal state of a session
// without an interactive desktop logon. A scope recorded on the strength of
// such a write looks configured and fails at the first fetch that needs it.
func (a Access) StoreScoped(ctx context.Context, scope, host, secret string) error {
	if !portableValue(scope) || !portableValue(host) {
		return errors.New("credential scope and host must be non-empty single lines")
	}
	if !portableValue(secret) {
		return errors.New("a credential must be a non-empty single line")
	}
	if _, ok := a.call(ctx, "approve",
		field{"protocol", "https"}, field{"host", host},
		field{"username", NamespaceUsername(scope)}, field{"password", secret}); !ok {
		return notPersistedError("your Git credential helper refused to store the credential")
	}
	if stored, ok := a.ReadScoped(ctx, scope, host); !ok || stored != secret {
		return notPersistedError("your Git credential helper reported success but did not persist the credential")
	}
	return nil
}

// DeleteScoped removes the manager-stored credential for scope on host and
// reports whether it is gone. The removal is proved the same way a write is:
// a helper that accepts a rejection it does not act on leaves the operator
// believing a credential was withdrawn.
func (a Access) DeleteScoped(ctx context.Context, scope, host string) bool {
	if !portableValue(scope) || !portableValue(host) {
		return false
	}
	if _, ok := a.call(ctx, "reject",
		field{"protocol", "https"}, field{"host", host},
		field{"username", NamespaceUsername(scope)}); !ok {
		return false
	}
	_, present := a.ReadScoped(ctx, scope, host)
	return !present
}

// Discover lists, without retaining, the operator material host could
// authenticate with. Only scopes addressed by that host are asked about; a
// scope belonging to another host has no entry to find here.
//
// This is the presence probe a prompt is built from: it reads exactly what a
// real resolution would read, and keeps none of it.
func (a Access) Discover(ctx context.Context, host string, scopes []string) HostMaterial {
	var material HostMaterial
	if own, ok := a.ReadHost(ctx, host); ok {
		material.HostCredential, material.HostUsername = true, own.Username
	}
	for _, scope := range scopes {
		if ScopeHost(scope) != host {
			continue
		}
		if _, ok := a.ReadScoped(ctx, scope, host); ok {
			material.Scopes = append(material.Scopes, scope)
		}
	}
	return material
}

// field is one line of the Git credential protocol. The order the caller
// writes them in is the order they are sent, so one request is one fixed
// payload.
type field struct{ key, value string }

// call runs one `git credential <action>` and returns the answer it printed.
//
// Anything that is not a clean, zero-status answer is reported as no answer:
// an absent Git, a helper that failed, a prompt that was refused. A read
// distinguishes none of them, because none of them is material.
func (a Access) call(ctx context.Context, action string, request ...field) (map[string]string, bool) {
	executable := a.Git
	if executable == "" {
		resolved, err := exec.LookPath("git")
		if err != nil {
			return nil, false
		}
		executable = resolved
	}
	timeout := a.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var payload strings.Builder
	for _, item := range request {
		payload.WriteString(item.key + "=" + item.value + "\n")
	}
	payload.WriteString("\n")

	// core.askPass is emptied on the command line as well as unset in the
	// environment: an operator askpass program is a prompt, and a prompt is
	// exactly what a manager-driven read must never reach.
	cmd := exec.CommandContext(ctx, executable, // #nosec G204 -- operator-selected Git executable, closed argument construction.
		"-c", "core.askPass=", "-c", "credential.interactive=false", "credential", action)
	answer := &boundedBuffer{remaining: maxAnswerBytes}
	cmd.Env, cmd.Stdin = a.environment(), strings.NewReader(payload.String())
	cmd.Stdout, cmd.Stderr = answer, io.Discard
	cmd.WaitDelay = drainDelay
	if err := cmd.Run(); err != nil {
		return nil, false
	}
	return parseAnswer(answer.Bytes()), true
}

// environment is the operator's own environment with prompting closed off and
// the operator home pinned.
//
// The environment is inherited rather than rebuilt: a credential helper is a
// session-bound program, and the desktop bus, keychain session and proxy
// settings it needs live nowhere else. What is overridden is only what must
// be: the home the helper is looked up through, and every door an interactive
// prompt could come through.
func (a Access) environment() []string {
	home := a.Home
	if home == "" {
		home = OperatorHome()
	}
	// A terminal prompt, a Git askpass program, an OpenSSH askpass program
	// and a credential manager's own interactivity are four separate ways for
	// a read to stop and wait for a human. The askpass variables are dropped
	// rather than emptied: Git reads an empty one as unset and falls through
	// to the next prompt source.
	suppressed := map[string]bool{
		"GIT_TERMINAL_PROMPT": true, "GCM_INTERACTIVE": true,
		"GIT_ASKPASS": true, "SSH_ASKPASS": true,
	}
	if home != "" {
		suppressed["HOME"], suppressed["USERPROFILE"] = true, true
	}
	base := a.Environ
	if base == nil {
		base = os.Environ()
	}
	environment := make([]string, 0, len(base)+4)
	for _, entry := range base {
		key, _, found := strings.Cut(entry, "=")
		// Windows environment names are case-insensitive, so an inherited
		// variable is matched the same way rather than surviving under
		// another spelling.
		if !found || suppressed[strings.ToUpper(key)] {
			continue
		}
		environment = append(environment, entry)
	}
	environment = append(environment, "GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=never")
	if home != "" {
		// The fetch owns a private HOME, and Windows resolves a home through
		// its own variable, so both are pinned to the operator's.
		environment = append(environment, "HOME="+home, "USERPROFILE="+home)
	}
	return environment
}

// OperatorHome is the home directory of the account the manager runs as,
// whose Git configuration selects the credential helper. Empty when the
// platform cannot name one, in which case the call inherits whatever home the
// environment carries.
func OperatorHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// parseAnswer reads the key=value lines a credential call prints. A line
// without a separator is not part of the protocol and is dropped.
func parseAnswer(output []byte) map[string]string {
	answer := map[string]string{}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if key, value, found := strings.Cut(line, "="); found {
			answer[key] = value
		}
	}
	return answer
}

// portableValue reports a value the credential protocol can carry. The
// protocol is newline-delimited, so a value carrying a line break or a NUL
// would not be one value; refusing it here is what keeps a request one
// request.
func portableValue(value string) bool {
	return value != "" && !strings.ContainsAny(value, "\n\r\x00")
}

// notPersistedError explains a write that left nothing behind, and names the
// store the operator would configure on the platform they are on. The
// guidance for every platform is kept in the message: a manager is routinely
// driven from one platform against notes written on another.
func notPersistedError(lead string) error {
	guidance := map[string]string{
		"darwin":  "on macOS run 'git config --global credential.helper osxkeychain'",
		"windows": "on Windows use an interactive logon session, or run 'git config --global credential.credentialStore dpapi'",
	}[runtime.GOOS]
	if guidance == "" {
		guidance = "on Linux run 'git config --global credential.helper libsecret'"
	}
	return fmt.Errorf("%s.\nConfigure a working credential store — %s "+
		"(other platforms: osxkeychain on macOS, libsecret on Linux, dpapi on Windows) "+
		"— or select an environment-variable credential source instead", lead, guidance)
}

// boundedBuffer collects a helper answer up to a fixed size. A helper is an
// operator-configured program, and a manager must not grow without bound on
// whatever it decides to print.
type boundedBuffer struct {
	buffer    bytes.Buffer
	remaining int
}

func (b *boundedBuffer) Write(payload []byte) (int, error) {
	accepted := payload
	if len(accepted) > b.remaining {
		accepted = accepted[:b.remaining]
	}
	if len(accepted) > 0 {
		if _, err := b.buffer.Write(accepted); err != nil {
			return 0, err
		}
		b.remaining -= len(accepted)
	}
	return len(payload), nil
}

func (b *boundedBuffer) Bytes() []byte { return b.buffer.Bytes() }
