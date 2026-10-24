package claudequalification

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// The intake control proves missing opt-in skips and a wrong digest is rejected
// before a fixture exists. The fabricated file is never accepted as a vendor.
func TestIntakeBoundary(t *testing.T) {
	p := setup(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrong := filepath.Join(p.root, "wrong-binary")
	write(t, wrong, "synthetic non-vendor bytes\n")
	for _, tc := range []struct{ name, binary, want string }{{"missing", "", "opt-in: CLAUDE_QUAL_BINARY unset"}, {"wrong-digest", wrong, "unqualified binary digest"}} {
		t.Run(tc.name, func(t *testing.T) {
			childRoot := filepath.Join(p.root, tc.name)
			env := []string{}
			for _, v := range os.Environ() {
				if !strings.HasPrefix(v, "CLAUDE_QUAL_BINARY=") && !strings.HasPrefix(v, "CLAUDE_QUAL_RUN_ROOT=") {
					env = append(env, v)
				}
			}
			env = append(env, "CLAUDE_QUAL_BINARY="+tc.binary, "CLAUDE_QUAL_RUN_ROOT="+childRoot)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, self, "-test.run=^TestIdentity$", "-test.v")
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
			if (err == nil) != (tc.binary == "") {
				t.Fatalf("unexpected child verdict: %v %s", err, out)
			}
			contains(t, string(out), tc.want, true)
			if _, err := os.Stat(filepath.Join(childRoot, "TestIdentity")); !os.IsNotExist(err) {
				t.Fatalf("intake created vendor fixture: %v", err)
			}
			p.save("intake-"+tc.name+".log", string(out))
		})
	}
}

// Synthetic boundary controls pair reachable provider and ordinary-file reads
// with denied unrelated loopback, synthetic Keychains bytes and outside writes.
// Mutants alter only these curl/cat/sh controls; no vendor runs under a mutant.
func TestOfflineBoundary(t *testing.T) {
	p := setup(t)
	var mu sync.Mutex
	hits := 0
	unrelated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		hits++
		mu.Unlock()
		_, _ = io.WriteString(w, "UNRELATED_ENDPOINT")
	}))
	defer unrelated.Close()
	policy := p.policy
	if mutant := os.Getenv("CLAUDE_QUAL_BOUNDARY_MUTANT"); mutant != "" {
		sb := read(t, p.policy)
		switch mutant {
		case "network":
			sb = strings.Replace(sb, "(deny network*)\n", "", 1)
		case "keychains":
			sb = strings.Replace(sb, "(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n", "", 1)
		default:
			t.Fatal("unknown boundary mutant")
		}
		policy = filepath.Join(p.root, "synthetic-control.sb")
		write(t, policy, sb)
	}
	command := func(name string, args ...string) (string, error) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		wrapped := append([]string{"-f", policy, name}, args...)
		cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", wrapped...)
		cmd.Env = p.env
		cmd.Dir = p.work
		out, err := cmd.CombinedOutput()
		code := 0
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				code = e.ExitCode()
			} else {
				t.Fatal(err)
			}
		}
		p.save(fmt.Sprintf("control-%d-%s.log", len(p.requests), filepath.Base(name)+"-"+strconv.Itoa(code)), fmt.Sprintf("command=%s %v\nexit=%d\n%s", name, args, code, out))
		if ctx.Err() != nil {
			t.Fatal("control timed out")
		}
		return string(out), err
	}
	out, err := command("/usr/bin/curl", "--silent", "--show-error", "--max-time", "2", "-H", "Content-Type: application/json", "--data", `{"model":"fixture","messages":[]}`, p.server.URL+"/v1/messages")
	if err != nil || !strings.Contains(out, "OFFLINE_OK") {
		t.Fatalf("allowed provider failed: %v %s", err, out)
	}
	out, err = command("/usr/bin/curl", "--silent", "--show-error", "--max-time", "2", unrelated.URL)
	if err == nil || !strings.Contains(out, "Failed to connect") {
		t.Fatalf("unrelated loopback not denied: %v %s", err, out)
	}
	mu.Lock()
	gotHits := hits
	mu.Unlock()
	if gotHits != 0 {
		t.Fatalf("unrelated network effects=%d", gotHits)
	}
	ordinary := filepath.Join(p.root, "ordinary.txt")
	blocked := filepath.Join(p.root, "Keychains", "synthetic.txt")
	for _, path := range []string{ordinary, blocked} {
		write(t, path, "SYNTHETIC_FILE_CONTROL\n")
	}
	out, err = command("/bin/cat", ordinary)
	if err != nil || out != "SYNTHETIC_FILE_CONTROL\n" {
		t.Fatalf("ordinary read: %v %s", err, out)
	}
	out, err = command("/bin/cat", blocked)
	if err == nil || !strings.Contains(out, "Operation not permitted") {
		t.Fatalf("synthetic Keychains read not denied: %v %s", err, out)
	}
	if read(t, blocked) != "SYNTHETIC_FILE_CONTROL\n" {
		t.Fatal("denied source changed")
	}
	outside := filepath.Join(filepath.Dir(p.root), "outside-write.txt")
	write(t, outside, "ORIGINAL\n")
	out, err = command("/bin/sh", "-c", "printf changed > \"$1\"", "probe", outside)
	if err == nil || !strings.Contains(out, "Operation not permitted") || read(t, outside) != "ORIGINAL\n" {
		t.Fatalf("outside write not rejected unchanged: %v %s", err, out)
	}
	inside := filepath.Join(p.root, "allowed-write.txt")
	out, err = command("/bin/sh", "-c", "printf changed > \"$1\"", "probe", inside)
	if err != nil || read(t, inside) != "changed" {
		t.Fatalf("inside write: %v %s", err, out)
	}
}
