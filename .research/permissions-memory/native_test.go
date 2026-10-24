package permissionsmemory

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const museSHA = "7118b9ca2227a6ec1b90cc506f8abdd4a35d14d8ab6683c9164b075002a55fff"
const codexSHA = "8813367285524ac798dd96d1b071aea30d58708c7226fd236cd275cf1cfdee31"

func (p *probe) nativeRun(label string, args ...string) (string, int) {
	p.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	env := append(append([]string{}, p.env...), "MUSE_NO_AUTO_UPDATE=1", "TBH_CREDENTIAL_BACKEND=file", "CODEX_HOME="+filepath.Join(p.root, "codex"))
	wrapped := append([]string{"-f", p.policy, "/usr/bin/env", "-i"}, env...)
	wrapped = append(wrapped, p.binary)
	wrapped = append(wrapped, args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", wrapped...)
	cmd.Stdin = strings.NewReader(p.nativeInput)
	cmd.Dir = p.work
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
		} else {
			p.t.Fatal(err)
		}
	}
	p.save(label+".json", mustJSON(p.t, map[string]any{"argv": args, "exit": code, "pid": cmd.Process.Pid, "output": string(out), "timeout": ctx.Err() != nil}))
	if ctx.Err() != nil {
		p.t.Fatal("native command timeout is unknown, never absence evidence")
	}
	return string(out), code
}

// Claim: exact-digest offline introspection records the real vendor version and
// help, then a Muse built-in echo baseline. Echo is no permission/memory proof:
// it requests no Bash/Edit/Write/WebFetch/MCP tool and cannot qualify denials.
func TestNativeIntrospection(t *testing.T) {
	if os.Getenv("QUAL_MUSE_BINARY") == "" && os.Getenv("QUAL_CODEX_BINARY") == "" {
		t.Skip("vendor opt-ins unset")
	}
	for _, runtime := range []string{"muse", "codex"} {
		t.Run(runtime, func(t *testing.T) {
			key, pin, want := "QUAL_MUSE_BINARY", museSHA, "1.4.12"
			if runtime == "codex" {
				key, pin, want = "QUAL_CODEX_BINARY", codexSHA, "0.163.0"
			}
			p := setupBinary(t, key, map[string]bool{pin: true})
			out, code := p.nativeRun("version", "--version")
			if code != 0 || !strings.Contains(out, want) {
				t.Fatal("pinned version missing")
			}
			data, _ := os.ReadFile(p.binary)
			p.save("identity.json", mustJSON(t, map[string]any{"runtime": runtime, "sha256": fmt.Sprintf("%x", sha256.Sum256(data)), "version": out, "platform": "darwin/arm64"}))
			p.nativeRun("help", "--help")
			p.nativeRun("exec-help", "exec", "--help")
			if runtime == "codex" {
				p.nativeRun("resume-help", "exec", "resume", "--help")
			} else {
				p.nativeRun("resume-help", "resume", "--help")
			}
			if runtime == "muse" {
				out, code = p.nativeRun("echo", "exec", "--json", "--provider", "echo", "--trust-workspace", "QUAL_ECHO_BASELINE")
				if code != 0 || !strings.Contains(out, "QUAL_ECHO_BASELINE") {
					t.Fatal("missing offline echo baseline")
				}
			}
		})
	}
}

// Claim: the same synthetic write succeeds inside the privacy fence directly;
// an all-allow nested Seatbelt profile either succeeds too or fails specifically
// at sandbox_apply without an effect. The latter isolates environment nesting
// from any Codex read-only/workspace policy; it is never a policy-denial proof.
func TestNestedSandboxControl(t *testing.T) {
	p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
	target := filepath.Join(p.work, "nested-effect")
	run := func(nested bool) (string, error) {
		args := []string{"-f", p.policy, "/bin/sh", "-c", "printf QUAL_NESTED > " + quote(target)}
		if nested {
			args = []string{"-f", p.policy, "/usr/bin/sandbox-exec", "-p", "(version 1)(allow default)", "/bin/sh", "-c", "printf QUAL_NESTED > " + quote(target)}
		}
		cmd := exec.Command("/usr/bin/sandbox-exec", args...)
		cmd.Dir = p.work
		cmd.Env = p.env
		out, err := cmd.CombinedOutput()
		p.save(fmt.Sprintf("nested-%t.json", nested), mustJSON(t, map[string]any{"output": string(out), "failed": err != nil, "pid": cmd.Process.Pid}))
		return string(out), err
	}
	if out, err := run(false); err != nil || read(t, target) != "QUAL_NESTED" {
		t.Fatalf("outer shell positive unavailable: %v %s", err, out)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	out, err := run(true)
	b, statErr := os.ReadFile(target)
	nestingDenied := err != nil && strings.Contains(out, "sandbox_apply: Operation not permitted") && os.IsNotExist(statErr)
	if !nestingDenied && (err != nil || string(b) != "QUAL_NESTED") {
		t.Fatalf("unrelated nested control failure: %v %s", err, out)
	}
	p.save("environment.json", mustJSON(t, map[string]any{"outerWrite": true, "allAllowNestedDenied": nestingDenied, "interpretation": "environment nesting only; not a manager policy verdict"}))
}
