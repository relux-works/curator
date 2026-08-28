package yarnmodernsource

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/nodesource"
)

const checksumA = "10c0/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const checksumB = "10c0/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func TestS01N01ModernLockConfigLinkerConditionsAndChecksums(t *testing.T) {
	graph, err := Parse(baseParseRequest(baseLock()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if graph.Layout.NodeLinker != "pnp" || graph.Layout.CacheKey != "10c0" || graph.Layout.CompressionLevel != 0 || graph.Layout.ConditionGrammar != ConditionGrammarID {
		t.Fatalf("layout = %+v", graph.Layout)
	}
	if len(graph.Packages) != 3 || len(graph.Edges) != 2 {
		t.Fatalf("packages/edges = %d/%d", len(graph.Packages), len(graph.Edges))
	}
	for _, pkg := range graph.Packages {
		if !pkg.Selected {
			t.Fatalf("package %q was pruned: %s", pkg.Key, pkg.PruneReason)
		}
	}
	if len(graph.Layout.BuiltinPlugins) != len(BuiltinPlugins) {
		t.Fatalf("builtins = %d", len(graph.Layout.BuiltinPlugins))
	}
}

func TestS06CanonicalModernLockIgnoresEntryOrder(t *testing.T) {
	first, err := Parse(baseParseRequest(baseLock()))
	if err != nil {
		t.Fatal(err)
	}
	secondLock := strings.Replace(baseLock(), remoteEntries(), reverseRemoteEntries(), 1)
	second, err := Parse(baseParseRequest(secondLock))
	if err != nil {
		t.Fatal(err)
	}
	if first.LockDigest != second.LockDigest {
		t.Fatalf("lock digests differ: %s != %s", first.LockDigest, second.LockDigest)
	}
}

func TestModernLockGrammarIsStrictTypedAndSingleDocument(t *testing.T) {
	baseline, err := Parse(baseParseRequest(baseLock()))
	if err != nil {
		t.Fatal(err)
	}
	variants := map[string]string{
		"second document":             baseLock() + "---\nignored: true\n",
		"metadata sequence":           strings.Replace(baseLock(), "__metadata:\n  version: 8\n  cacheKey: 10c0", "__metadata: []", 1),
		"metadata version string":     strings.Replace(baseLock(), "  version: 8\n", "  version: \"8\"\n", 1),
		"metadata cache key integer":  strings.Replace(baseLock(), "  cacheKey: 10c0\n", "  cacheKey: 10\n", 1),
		"metadata unknown field":      strings.Replace(baseLock(), "  cacheKey: 10c0\n", "  cacheKey: 10c0\n  futureBehavior: true\n", 1),
		"entry sequence":              strings.Replace(baseLock(), "\"a@npm:^1.0.0\":\n  version:", "\"a@npm:^1.0.0\": []\n\"ignored@npm:1.0.0\":\n  version:", 1),
		"version sequence":            strings.Replace(baseLock(), "  version: 1.0.0\n  resolution: \"a@npm:1.0.0\"", "  version: []\n  resolution: \"a@npm:1.0.0\"", 1),
		"resolution mapping":          strings.Replace(baseLock(), "  resolution: \"a@npm:1.0.0\"", "  resolution: {}", 1),
		"checksum boolean":            strings.Replace(baseLock(), "  checksum: \""+checksumA+"\"", "  checksum: true", 1),
		"language sequence":           strings.Replace(baseLock(), "  languageName: node", "  languageName: [node]", 1),
		"link mapping":                strings.Replace(baseLock(), "  linkType: hard", "  linkType: {kind: hard}", 1),
		"dependencies sequence":       strings.Replace(baseLock(), "  dependencies:\n    b: \"npm:1.0.0\"", "  dependencies: []", 1),
		"dependency value boolean":    strings.Replace(baseLock(), "    b: \"npm:1.0.0\"", "    b: true", 1),
		"peers scalar":                strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  peerDependencies: react\n  conditions: \"os=linux & cpu=x64\"", 1),
		"dependency metadata scalar":  strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  dependenciesMeta: false\n  conditions: \"os=linux & cpu=x64\"", 1),
		"metadata entry scalar":       strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  dependenciesMeta:\n    b: true\n  conditions: \"os=linux & cpu=x64\"", 1),
		"metadata optional string":    strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  dependenciesMeta:\n    b:\n      optional: \"true\"\n  conditions: \"os=linux & cpu=x64\"", 1),
		"metadata unknown nested key": strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  dependenciesMeta:\n    b:\n      optional: true\n      built: false\n  conditions: \"os=linux & cpu=x64\"", 1),
		"conditions mapping":          strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  conditions: {os: linux}", 1),
		"conditions mixed sequence":   strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  conditions: [\"os=linux\", 7]", 1),
		"duplicate entry field":       strings.Replace(baseLock(), "  checksum: \""+checksumA+"\"", "  checksum: \""+checksumA+"\"\n  checksum: \""+checksumA+"\"", 1),
	}
	for name, lock := range variants {
		t.Run(name, func(t *testing.T) {
			rejected, parseErr := Parse(baseParseRequest(lock))
			if ErrorCode(parseErr) != CodeLockFormatUnsupported {
				t.Fatalf("error = %v (%s), want %s", parseErr, ErrorCode(parseErr), CodeLockFormatUnsupported)
			}
			if rejected.LockDigest != "" || rejected.RawLockSHA256 != "" || rejected.ConfigurationDigest != "" || len(rejected.Packages) != 0 || len(rejected.Edges) != 0 {
				t.Fatalf("rejected lock emitted graph/config identity or aliased %s: %+v", baseline.LockDigest, rejected)
			}
		})
	}
}

func TestModernLockCanonicalIdentityBindsOptionalMetadata(t *testing.T) {
	baseline, err := Parse(baseParseRequest(baseLock()))
	if err != nil {
		t.Fatal(err)
	}
	withMetadata := strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"", "  dependenciesMeta:\n    b:\n      optional: true\n  conditions: \"os=linux & cpu=x64\"", 1)
	variant, err := Parse(baseParseRequest(withMetadata))
	if err != nil {
		t.Fatal(err)
	}
	if baseline.LockDigest == variant.LockDigest {
		t.Fatalf("behavior-affecting dependency metadata aliased lock identity %s", baseline.LockDigest)
	}
}

func TestN02N10LockRuntimeAndTargetDriftFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ParseRequest)
		code   string
	}{
		{"release", func(r *ParseRequest) { r.YarnVersion = "4.9.1" }, CodeLockFormatUnsupported},
		{"packageManager", func(r *ParseRequest) {
			r.Manifests["package.json"] = []byte(`{"name":"root","version":"1.0.0","packageManager":"yarn@4.9.1","dependencies":{"a":"npm:^1.0.0"}}`)
		}, CodeLockStale},
		{"target", func(r *ParseRequest) { r.Target.OS = "darwin" }, CodeGraphIncomplete},
		{"checksum", func(r *ParseRequest) { r.LockBytes = []byte(strings.Replace(baseLock(), checksumA, "11/abcd", 1)) }, CodeIntegrityMissing},
		{"workspace link identity", func(r *ParseRequest) {
			r.LockBytes = []byte(strings.Replace(baseLock(), "  languageName: unknown\n  linkType: soft\n", "  languageName: node\n  linkType: hard\n", 1))
		}, CodeLockFormatUnsupported},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := baseParseRequest(baseLock())
			test.mutate(&request)
			_, err := Parse(request)
			if ErrorCode(err) != test.code {
				t.Fatalf("error = %v (%s), want %s", err, ErrorCode(err), test.code)
			}
		})
	}
}

func TestN01N02WorkspaceLockEntriesReconcileExactly(t *testing.T) {
	request := workspaceParseRequest()
	graph, err := Parse(request)
	if err != nil {
		t.Fatalf("Parse() positive workspace error = %v", err)
	}
	if len(graph.Workspaces) != 1 || len(graph.Packages) != 2 || len(graph.Edges) != 1 || graph.Packages[graph.packageIndex["workspace:packages/child"]].Version != "2.3.4" {
		t.Fatalf("workspace graph = workspaces:%v packages:%+v edges:%+v", graph.Workspaces, graph.Packages, graph.Edges)
	}

	tests := map[string]func(*ParseRequest){
		"missing root entry": func(r *ParseRequest) {
			child := strings.Index(string(r.LockBytes), "\"child@workspace:")
			r.LockBytes = append([]byte("__metadata:\n  version: 8\n  cacheKey: 10c0\n"), r.LockBytes[child:]...)
		},
		"missing child entry": func(r *ParseRequest) {
			r.LockBytes = r.LockBytes[:strings.Index(string(r.LockBytes), "\"child@workspace:")]
		},
		"workspace path drift": func(r *ParseRequest) {
			r.LockBytes = []byte(strings.ReplaceAll(string(r.LockBytes), "workspace:packages/child", "workspace:packages/other"))
		},
		"workspace name drift": func(r *ParseRequest) {
			r.LockBytes = []byte(strings.ReplaceAll(string(r.LockBytes), "child@workspace", "renamed@workspace"))
		},
		"workspace version drift": func(r *ParseRequest) {
			r.LockBytes = []byte(strings.Replace(string(r.LockBytes), "version: 0.0.0-use.local", "version: 2.3.4", 1))
		},
		"workspace selector drift": func(r *ParseRequest) {
			r.LockBytes = []byte(strings.Replace(string(r.LockBytes), "child@workspace:*, child@workspace:packages/child", "child@workspace:^, child@workspace:packages/child", 1))
		},
		"duplicate workspace entry": func(r *ParseRequest) {
			r.LockBytes = append(r.LockBytes, []byte("\"child@workspace:duplicate\":\n  version: 0.0.0-use.local\n  resolution: \"child@workspace:packages/child\"\n  languageName: unknown\n  linkType: soft\n")...)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			variant := workspaceParseRequest()
			mutate(&variant)
			rejected, err := Parse(variant)
			if ErrorCode(err) != CodeLockStale {
				t.Fatalf("error = %v (%s), want %s", err, ErrorCode(err), CodeLockStale)
			}
			if len(rejected.Packages) != 0 || len(rejected.Edges) != 0 || rejected.ConfigurationDigest != "" {
				t.Fatalf("rejected workspace drift emitted graph evidence: %+v", rejected)
			}
		})
	}
}

func TestN02WorkspaceDependencyScopesReconcileBeforeExecution(t *testing.T) {
	tests := []struct {
		name         string
		manifestPart string
		lockPart     string
		mutate       func(string) string
	}{
		{"dependency", `"dependencies":{"a":"npm:^1.0.0"}`, "  dependencies:\n    a: \"npm:^1.0.0\"\n", func(lock string) string {
			return strings.Replace(lock, "  dependencies:\n    a: \"npm:^1.0.0\"\n", "", 1)
		}},
		{"development", `"devDependencies":{"a":"npm:^1.0.0"}`, "  dependencies:\n    a: \"npm:^1.0.0\"\n", func(lock string) string { return strings.Replace(lock, "npm:^1.0.0", "npm:1.0.0", 1) }},
		{"optional", `"optionalDependencies":{"a":"npm:^1.0.0"}`, "  dependencies:\n    a: \"npm:^1.0.0\"\n  dependenciesMeta:\n    a:\n      optional: true\n", func(lock string) string {
			return strings.Replace(lock, "  dependenciesMeta:\n    a:\n      optional: true\n", "", 1)
		}},
		{"peer", `"peerDependencies":{"a":"npm:^1.0.0"},"peerDependenciesMeta":{"a":{"optional":true}}`, "  peerDependencies:\n    a: \"npm:^1.0.0\"\n  peerDependenciesMeta:\n    a:\n      optional: true\n", func(lock string) string {
			return strings.Replace(lock, "  peerDependenciesMeta:\n    a:\n      optional: true\n", "", 1)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2",` + test.manifestPart + `}`)
			lock := "__metadata:\n  version: 8\n  cacheKey: 10c0\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n" + test.lockPart + "  languageName: unknown\n  linkType: soft\n" + remoteEntries()
			request := baseParseRequest(lock)
			request.Manifests["package.json"] = manifest
			if _, err := Parse(request); err != nil {
				t.Fatalf("positive scope Parse() error = %v", err)
			}
			request.LockBytes = []byte(test.mutate(lock))
			rejected, err := Parse(request)
			if ErrorCode(err) != CodeLockStale {
				t.Fatalf("drift error = %v (%s), want %s", err, ErrorCode(err), CodeLockStale)
			}
			if len(rejected.Packages) != 0 || len(rejected.Edges) != 0 || rejected.ConfigurationDigest != "" {
				t.Fatalf("rejected workspace metadata emitted graph evidence: %+v", rejected)
			}
		})
	}
}

func TestN10YarnConditionGrammarMatchesPinnedRelease(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		target     Target
		matched    bool
	}{
		{"or selects linux", "(os=linux | os=darwin)", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"or prunes windows", "(os=linux | os=darwin)", Target{OS: "win32", Architecture: "x64", Libc: "glibc"}, false},
		{"negated selector", "!os=darwin", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"negated selector prunes", "!os=linux", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, false},
		{"even repeated negation", "!!os=linux", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"odd repeated negation", "!!!os=darwin", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"single-selector group", "(os=linux)", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"ungrouped or", "os=darwin | os=linux", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"xor", "os=linux ^ cpu=arm64", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
		{"conjunction", "(os=linux | os=darwin) & cpu=x64 & !libc=musl", Target{OS: "linux", Architecture: "x64", Libc: "glibc"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matched, _, err := evaluateYarnCondition(test.expression, map[string]string{"os": test.target.OS, "cpu": test.target.Architecture, "libc": test.target.Libc})
			if err != nil || matched != test.matched {
				t.Fatalf("evaluateYarnCondition(%q) = %v, %v; want %v", test.expression, matched, err, test.matched)
			}
		})
	}
	for _, malformed := range []string{"", "arch=x64", "os=linux && cpu=x64", "os=linux || cpu=x64", "os=Linux", "os=linux ", "(os=linux"} {
		if _, _, err := evaluateYarnCondition(malformed, map[string]string{"os": "linux", "cpu": "x64", "libc": "glibc"}); err == nil {
			t.Fatalf("malformed condition %q admitted", malformed)
		}
	}

	request := baseParseRequest(strings.Replace(baseLock(), "os=linux & cpu=x64", "(os=linux | os=darwin) & cpu=x64 & !libc=musl", 1))
	graph, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	condition := targetCondition(graph.Packages[graph.packageIndex["a@npm:1.0.0"]])
	if condition == nil || condition.Expression != "(os=linux | os=darwin) & cpu=x64 & !libc=musl" {
		t.Fatalf("capture condition = %+v", condition)
	}
	matched, err := evaluateYarnPlatformCondition(*condition, closuregraph.EvaluationInput{Selection: closuregraph.SelectionContext{Markers: map[string]string{"os": "linux", "cpu": "x64", "libc": "glibc"}}})
	if err != nil || !matched {
		t.Fatalf("C4 condition = %v, %v", matched, err)
	}
	optionalLock := strings.Replace(baseLock(), "  dependencies:\n    a: \"npm:^1.0.0\"\n", "  dependencies:\n    a: \"npm:^1.0.0\"\n  dependenciesMeta:\n    a:\n      optional: true\n", 1)
	optionalLock = strings.Replace(optionalLock, "os=linux & cpu=x64", "os=darwin", 1)
	optionalRequest := baseParseRequest(optionalLock)
	optionalRequest.Manifests["package.json"] = []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","optionalDependencies":{"a":"npm:^1.0.0"}}`)
	pruned, err := Parse(optionalRequest)
	if err != nil {
		t.Fatal(err)
	}
	a := pruned.Packages[pruned.packageIndex["a@npm:1.0.0"]]
	var rootEdge DependencyEdge
	for _, edge := range pruned.Edges {
		if edge.From == "workspace:." && edge.To == "a@npm:1.0.0" {
			rootEdge = edge
		}
	}
	if a.Selected || a.PruneReason != "os_pruned" || rootEdge.Selected || rootEdge.Reason != "os_pruned" {
		t.Fatalf("pruned package/edge = %+v / %+v", a, rootEdge)
	}
}

func TestN10MalformedConditionsFailBeforeGraphIdentityAndDownstreamWork(t *testing.T) {
	optionalLock := strings.Replace(baseLock(), "  dependencies:\n    a: \"npm:^1.0.0\"\n", "  dependencies:\n    a: \"npm:^1.0.0\"\n  dependenciesMeta:\n    a:\n      optional: true\n", 1)
	optionalLock = strings.Replace(optionalLock, "os=linux & cpu=x64", "os=linux && cpu=x64", 1)
	optional := baseParseRequest(optionalLock)
	optional.Manifests["package.json"] = []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","optionalDependencies":{"a":"npm:^1.0.0"}}`)

	optionalPeerLock := strings.Replace(baseLock(), "  dependencies:\n    b: \"npm:1.0.0\"\n", "  peerDependencies:\n    b: \"npm:1.0.0\"\n  peerDependenciesMeta:\n    b:\n      optional: true\n", 1)
	optionalPeerLock = strings.Replace(optionalPeerLock, "\"b@npm:1.0.0\":\n  version:", "\"b@npm:1.0.0\":\n  conditions: \"os=linux && cpu=x64\"\n  version:", 1)

	unreachableLock := baseLock() + "\"unreachable@npm:1.0.0\":\n  version: 1.0.0\n  resolution: \"unreachable@npm:1.0.0\"\n  checksum: \"10c0/cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc\"\n  conditions: \"os=linux && cpu=x64\"\n  languageName: node\n  linkType: hard\n"

	for name, request := range map[string]ParseRequest{
		"optional":      optional,
		"optional peer": baseParseRequest(optionalPeerLock),
		"unreachable":   baseParseRequest(unreachableLock),
	} {
		t.Run(name, func(t *testing.T) {
			graph, err := Parse(request)
			if ErrorCode(err) != CodeLockFormatUnsupported {
				t.Fatalf("Parse() error = %v (%s), want %s", err, ErrorCode(err), CodeLockFormatUnsupported)
			}
			if graph.LockDigest != "" || graph.RawLockSHA256 != "" || graph.ConfigurationDigest != "" || len(graph.Packages) != 0 || len(graph.Edges) != 0 || len(graph.Patches) != 0 {
				t.Fatalf("rejected condition emitted identity usable by capture, manager, build, cache, or publication: %+v", graph)
			}
		})
	}
}

func TestN10RepeatedNegationSelectionMatchesPinnedYarn(t *testing.T) {
	tests := []struct {
		expression string
		selected   bool
		reason     string
	}{
		{expression: "!!os=linux", selected: true},
		{expression: "!!!os=darwin", selected: true},
		{expression: "!!!os=linux", selected: false, reason: "os_pruned"},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			lock := strings.Replace(baseLock(), "  dependencies:\n    a: \"npm:^1.0.0\"\n", "  dependencies:\n    a: \"npm:^1.0.0\"\n  dependenciesMeta:\n    a:\n      optional: true\n", 1)
			lock = strings.Replace(lock, "os=linux & cpu=x64", test.expression, 1)
			request := baseParseRequest(lock)
			request.Manifests["package.json"] = []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","optionalDependencies":{"a":"npm:^1.0.0"}}`)
			graph, err := Parse(request)
			if err != nil {
				t.Fatal(err)
			}
			pkg := graph.Packages[graph.packageIndex["a@npm:1.0.0"]]
			if pkg.Selected != test.selected || pkg.PruneReason != test.reason {
				t.Fatalf("selection = %v reason = %q, want %v/%q", pkg.Selected, pkg.PruneReason, test.selected, test.reason)
			}
		})
	}
}

func TestN10RequiredPeerMustResolveAndOptionalPeerIsExplicitlyPruned(t *testing.T) {
	required := baseParseRequest(strings.Replace(baseLock(), "  conditions: \"os=linux & cpu=x64\"\n", "  peerDependencies:\n    react: \"npm:^18.0.0\"\n  conditions: \"os=linux & cpu=x64\"\n", 1))
	_, err := Parse(required)
	if ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("required peer error = %v", err)
	}

	optional := required
	optional.LockBytes = []byte(strings.Replace(string(required.LockBytes), "  conditions: \"os=linux & cpu=x64\"\n", "  peerDependenciesMeta:\n    react:\n      optional: true\n  conditions: \"os=linux & cpu=x64\"\n", 1))
	graph, err := Parse(optional)
	if err != nil {
		t.Fatal(err)
	}
	var peer DependencyEdge
	for _, edge := range graph.Edges {
		if strings.HasPrefix(edge.From, "a@npm:1.0.0#peer:") && edge.Name == "react" {
			peer = edge
		}
	}
	if peer.Scope != "peer" || peer.To != "" || peer.Selected || peer.Reason != "optional_peer_unresolved" {
		t.Fatalf("optional peer evidence = %+v", peer)
	}
}

func TestModernWorkspaceRuntimeCycleIsRetained(t *testing.T) {
	fixture := newModernWorkspaceCycleFixture(t)

	wantEdges := map[string]bool{
		"workspace:packages/a\x00b\x00workspace:packages/b": false,
		"workspace:packages/b\x00a\x00workspace:packages/a": false,
	}
	for _, edge := range fixture.graph.Edges {
		key := edge.From + "\x00" + edge.Name + "\x00" + edge.To
		if _, ok := wantEdges[key]; ok && edge.Scope == "runtime" && edge.Selected {
			wantEdges[key] = true
		}
	}
	for edge, found := range wantEdges {
		if !found {
			t.Errorf("selected runtime SCC edge %q is absent from canonical graph: %+v", edge, fixture.graph.Edges)
		}
	}
}

func TestModernRuntimeCyclesCaptureMaterializeAndInvoke(t *testing.T) {
	fixtures := map[string]func(*testing.T) modernExecutionFixture{
		"workspace":              newModernWorkspaceCycleFixture,
		"remote":                 newModernRemoteCycleFixture,
		"workspace peer context": newModernWorkspacePeerCycleFixture,
	}
	for name, makeFixture := range fixtures {
		t.Run(name, func(t *testing.T) {
			fixture := makeFixture(t)
			capture := captureModernFixture(t, fixture)
			runner := &modernFakeRunner{capture: capture}
			runner.authority = makeModernExecutionContext(t, capture, runner)
			cacheRoot := filepath.Join(t.TempDir(), "private-cache")
			cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), runner.authority)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
			materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, runner.authority)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = Invoke(t.Context(), materialized, "index.js", nil, runner.authority); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestModernRuntimeCyclesThroughRealPinnedYarn(t *testing.T) {
	yarnJS := os.Getenv("CURATOR_TEST_YARN_MODERN_JS")
	if yarnJS == "" {
		t.Skip("set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool")
	}
	fixtures := map[string]func(*testing.T) modernExecutionFixture{
		"workspace":              newModernWorkspaceCycleFixture,
		"remote":                 newModernRemoteCycleFixture,
		"workspace peer context": newModernWorkspacePeerCycleFixture,
	}
	for name, makeFixture := range fixtures {
		t.Run(name, func(t *testing.T) {
			fixture := makeFixture(t)
			capture := captureModernFixture(t, fixture)
			runner := newModernConcreteRunner(t, yarnJS)
			provider := newModernVerifiedProvider(runner)
			authority := makeModernVerifiedExecutionContext(t, capture, runner, provider)
			cacheRoot := filepath.Join(t.TempDir(), "private-cache")
			cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), authority)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
			materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, authority)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = Invoke(t.Context(), materialized, "index.js", nil, authority); err != nil {
				t.Fatal(err)
			}
			if provider.starts != 2 || len(runner.launches) != 2 {
				t.Fatalf("cycle execution launches = starts:%d launches:%+v", provider.starts, runner.launches)
			}
		})
	}
}

func TestModernPeerContextCycleMustReachTheSameDerivedInstance(t *testing.T) {
	base := []Package{
		{Key: "workspace:.", Name: "root", Version: "1.0.0", WorkspacePath: ".", Dependencies: map[string]string{"a": "npm:1.0.0", "host": "npm:1.0.0"}},
		{Key: "a@npm:1.0.0", Name: "a", Version: "1.0.0", Dependencies: map[string]string{"b": "npm:1.0.0"}, PeerDependencies: map[string]string{"host": "*"}},
		{Key: "b@npm:1.0.0", Name: "b", Version: "1.0.0", Dependencies: map[string]string{"a": "npm:1.0.0", "host": "npm:2.0.0"}},
		{Key: "host@npm:1.0.0", Name: "host", Version: "1.0.0"},
		{Key: "host@npm:2.0.0", Name: "host", Version: "2.0.0"},
	}
	selectors := map[string]string{
		"a@npm:1.0.0":    "a@npm:1.0.0",
		"b@npm:1.0.0":    "b@npm:1.0.0",
		"host@npm:1.0.0": "host@npm:1.0.0",
		"host@npm:2.0.0": "host@npm:2.0.0",
	}

	packages, edges, _, err := buildPackageGraph(base, selectors, map[string]string{})
	if ErrorCode(err) != CodeGraphIncomplete || !strings.Contains(err.Error(), "non-well-founded context cycle") {
		t.Fatalf("recursive peer context error = %v (%s), packages=%+v edges=%+v", err, ErrorCode(err), packages, edges)
	}
}

func TestModernPeerVersionGrammarIsClosed(t *testing.T) {
	tests := []struct {
		version, spec string
		want          bool
	}{
		{"1.2.3", "*", true}, {"1.2.3", "1.2.3", true}, {"1.2.3", "npm:^1.0.0", true},
		{"2.0.0", "^1.0.0", false}, {"0.2.9", "^0.2.0", true}, {"0.3.0", "^0.2.0", false},
		{"0.0.4", "^0.0.4", true}, {"0.0.5", "^0.0.4", false}, {"1.2.9", "~1.2.0", true},
		{"1.3.0", "~1.2.0", false}, {"1.2.3", ">=1.2.3", true}, {"1.2.3", ">1.2.3", false},
		{"1.2.3", "<=1.2.3", true}, {"1.2.3", "<1.2.3", false}, {"1.2.3", "^1 || ^2", false},
		{"1.2.3-beta.1", "^1.2.0", false}, {"1.2", "^1.2.0", false}, {"01.2.3", "^1.2.0", false},
	}
	for _, test := range tests {
		if got := modernPeerVersionSatisfies(test.version, test.spec); got != test.want {
			t.Errorf("modernPeerVersionSatisfies(%q, %q) = %v, want %v", test.version, test.spec, got, test.want)
		}
	}

	incompatible := newModernPeerWorkspaceFixture(t)
	lockBytes, err := os.ReadFile(filepath.Join(incompatible.root, "yarn.lock"))
	if err != nil {
		t.Fatal(err)
	}
	plugin := incompatible.graph.manifestBytes["packages/plugin/package.json"]
	incompatibleRequest := ParseRequest{LockName: "yarn.lock", LockBytes: []byte(strings.Replace(string(lockBytes), "host: \"*\"", "host: \"^2.0.0\"", 1)), Manifests: map[string][]byte{"package.json": incompatible.graph.manifestBytes["package.json"], "packages/host/package.json": incompatible.graph.manifestBytes["packages/host/package.json"], "packages/plugin/package.json": bytes.Replace(plugin, []byte(`"host":"*"`), []byte(`"host":"^2.0.0"`), 1)}, Configuration: map[string][]byte{".yarnrc.yml": incompatible.graph.configurationBytes[".yarnrc.yml"]}, YarnVersion: SupportedYarnVersion, Target: incompatible.graph.Target}
	if _, err = Parse(incompatibleRequest); ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("incompatible peer provider error = %v (%s)", err, ErrorCode(err))
	}
}

func TestN10WorkspacePeerContextIsDistinctAndPnPBijective(t *testing.T) {
	fixture := newModernPeerWorkspaceFixture(t)
	virtuals := []Package{}
	for _, pkg := range fixture.graph.Packages {
		if pkg.BaseKey != "" {
			virtuals = append(virtuals, pkg)
		}
	}
	if fixture.graph.Layout.PeerVirtualization != PeerVirtualizationAlgorithmID || len(virtuals) != 1 {
		t.Fatalf("peer virtualization = %q packages=%+v", fixture.graph.Layout.PeerVirtualization, virtuals)
	}
	virtual := virtuals[0]
	if virtual.BaseKey != "workspace:packages/plugin" || !virtual.Selected || len(virtual.PeerContext) != 2 {
		t.Fatalf("derived plugin context = %+v", virtual)
	}
	if virtual.PeerContext[0].Name != "@types/host" || !virtual.PeerContext[0].Optional || virtual.PeerContext[0].Provider != "" || virtual.PeerContext[1].Name != "host" || virtual.PeerContext[1].Provider != "workspace:packages/host" {
		t.Fatalf("exact provider context = %+v", virtual.PeerContext)
	}
	capture := captureModernFixture(t, fixture)
	loader, err := fixturePnPLoader(capture)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := reconcilePnPRuntimeState(loader, capture.Graph, capture.tarballs)
	if err != nil {
		t.Fatalf("derived virtual PnP state did not reconcile: %v", err)
	}
	if !containsStrings(selected, virtual.Key, "workspace:packages/plugin", "workspace:packages/host", "workspace:.") {
		t.Fatalf("selected contexts = %v", selected)
	}
	retargeted := bytes.Replace(loader, []byte("workspace:packages/host"), []byte("workspace:packages/plugin"), 1)
	if _, err = reconcilePnPRuntimeState(retargeted, capture.Graph, capture.tarballs); ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("retargeted virtual state error = %v (%s)", err, ErrorCode(err))
	}
}

func TestN10RemoteTransitivePeerContextsDoNotAlias(t *testing.T) {
	fixture := newModernTwoRemotePeerContextsFixture(t)
	providers := map[string]bool{}
	virtualKeys := []string{}
	for _, pkg := range fixture.graph.Packages {
		if pkg.BaseKey != "plugin@npm:1.0.0" {
			continue
		}
		virtualKeys = append(virtualKeys, pkg.Key)
		for _, binding := range pkg.PeerContext {
			if binding.Name == "host" {
				providers[binding.Provider] = true
			}
		}
	}
	if len(virtualKeys) != 2 || !providers["host@npm:1.0.0"] || !providers["host@npm:2.0.0"] || virtualKeys[0] == virtualKeys[1] {
		t.Fatalf("remote peer contexts aliased: keys=%v providers=%v", virtualKeys, providers)
	}
	capture := captureModernFixture(t, fixture)
	loader, err := fixturePnPLoader(capture)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = reconcilePnPRuntimeState(loader, capture.Graph, capture.tarballs); err != nil {
		t.Fatalf("remote transitive peer contexts did not reconcile: %v", err)
	}
	crossWired := bytes.Replace(loader, []byte(`npm:1.0.0`), []byte(`npm:2.0.0`), 1)
	if bytes.Equal(crossWired, loader) {
		t.Fatal("remote peer fixture did not expose the first host context")
	}
	if _, err = reconcilePnPRuntimeState(crossWired, capture.Graph, capture.tarballs); ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("cross-wired remote peer state error = %v (%s)", err, ErrorCode(err))
	}
}

func TestYarnVirtualRuntimeEncodingIsStrict(t *testing.T) {
	fixture := newModernPeerWorkspaceFixture(t)
	var virtual Package
	for _, pkg := range fixture.graph.Packages {
		if pkg.BaseKey != "" {
			virtual = pkg
			break
		}
	}
	locator, err := fixturePnPPackageLocator(virtual)
	if err != nil {
		t.Fatal(err)
	}
	base, err := yarnVirtualBaseLocator(locator)
	if err != nil || base != "plugin\x00workspace:packages/plugin" {
		t.Fatalf("virtual base = %q error=%v", base, err)
	}
	slug, err := yarnVirtualLocatorSlug(locator)
	if err != nil || !strings.HasPrefix(slug, "plugin-virtual-") || len(strings.TrimPrefix(slug, "plugin-virtual-")) != 10 {
		t.Fatalf("virtual slug = %q error=%v", slug, err)
	}
	hash := strings.Repeat("a", sha512.Size*2)
	scopedSlug, err := yarnVirtualLocatorSlug("@scope/pkg\x00virtual:" + hash + "#npm:1.0.0")
	if err != nil || !strings.HasPrefix(scopedSlug, "@scope-pkg-virtual-") {
		t.Fatalf("scoped virtual slug = %q error=%v", scopedSlug, err)
	}
	for _, malformed := range []string{"plugin\x00workspace:packages/plugin", "plugin\x00virtual:abcd", "plugin\x00virtual:" + strings.ToUpper(hash) + "#workspace:packages/plugin", "@broken\x00virtual:" + hash + "#npm:1.0.0"} {
		if _, err = yarnVirtualBaseLocator(malformed); err == nil && !strings.HasPrefix(malformed, "@broken") {
			t.Errorf("malformed virtual locator admitted: %q", malformed)
		}
		if strings.HasPrefix(malformed, "@broken") {
			if _, err = yarnVirtualLocatorSlug(malformed); err == nil {
				t.Errorf("malformed scoped locator admitted: %q", malformed)
			}
		}
	}
	if _, err = expectedPnPBaseLocation(Package{Key: "missing@npm:1.0.0", Name: "missing", Resolution: "missing@npm:1.0.0"}, map[string]capturedInput{}); ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("missing virtual cache authority error = %v (%s)", err, ErrorCode(err))
	}
}

func TestRawModernYarnTarballInspection(t *testing.T) {
	payload := testTGZ(t, map[string][]byte{
		"package.json": []byte(`{"name":"raw-package","version":"1.2.3"}`),
		"index.js":     []byte("module.exports = true\n"),
		"binding.gyp":  []byte("{}\n"),
	})
	manifest, inspection, files, err := inspectTarball(payload)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "raw-package" || manifest.Version != "1.2.3" || !inspection.bindingGYP || len(files) != 3 {
		t.Fatalf("raw tarball evidence = manifest:%+v inspection:%+v files:%+v", manifest, inspection, files)
	}
	if _, _, _, err = inspectTarball([]byte("not-gzip")); ErrorCode(err) != CodeMetadataMismatch {
		t.Fatalf("invalid raw tarball error = %v (%s)", err, ErrorCode(err))
	}
}

func TestModernRCBehaviorAffectingSettingsAreClosedAndBound(t *testing.T) {
	base, err := Parse(baseParseRequest(baseLock()))
	if err != nil {
		t.Fatal(err)
	}
	if base.Layout.EnableTelemetry || base.Layout.PnpEnableEsmLoader || base.Layout.DefaultProtocol != "npm:" || base.Layout.NpmRegistryServer != "https://registry.yarnpkg.com" {
		t.Fatalf("effective rc layout = %+v", base.Layout)
	}
	accepted := baseParseRequest(baseLock())
	accepted.Configuration[".yarnrc.yml"] = []byte(baseRC() + "enableTelemetry: false\npnpEnableEsmLoader: false\ndefaultProtocol: 'npm:'\nnpmRegistryServer: 'https://registry.yarnpkg.com'\n")
	explicit, err := Parse(accepted)
	if err != nil {
		t.Fatal(err)
	}
	if explicit.ConfigurationDigest != base.ConfigurationDigest {
		t.Fatalf("equivalent effective rc identities differ: %s != %s", explicit.ConfigurationDigest, base.ConfigurationDigest)
	}
	tests := map[string]string{
		"telemetry":        "enableTelemetry: true\n",
		"esm loader":       "pnpEnableEsmLoader: true\n",
		"default protocol": "defaultProtocol: 'file:'\n",
		"registry":         "npmRegistryServer: 'https://registry.example.test'\n",
	}
	for name, setting := range tests {
		t.Run(name, func(t *testing.T) {
			request := baseParseRequest(baseLock())
			request.Configuration[".yarnrc.yml"] = []byte(baseRC() + setting)
			_, err := Parse(request)
			if ErrorCode(err) != CodeManagerPluginUndeclared {
				t.Fatalf("error = %v", err)
			}
		})
	}
	malformed := map[string]string{
		"quoted telemetry boolean": "enableTelemetry: \"true\"\n",
		"scalar esm boolean":       "pnpEnableEsmLoader: nope\n",
		"quoted compression":       strings.Replace(baseRC(), "compressionLevel: 0", "compressionLevel: \"0\"", 1),
		"unknown architecture key": "supportedArchitectures:\n  invented: [linux]\n",
		"non-string architecture":  "supportedArchitectures:\n  os: [linux, 7]\n",
		"ambient current selector": "supportedArchitectures:\n  os: [current]\n",
		"duplicate selector":       "supportedArchitectures:\n  os: [linux, linux]\n",
		"duplicate setting":        "enableTelemetry: false\nenableTelemetry: false\n",
		"second yaml document":     "enableTelemetry: false\n---\nenableTelemetry: false\n",
	}
	for name, setting := range malformed {
		t.Run(name, func(t *testing.T) {
			request := baseParseRequest(baseLock())
			switch {
			case name == "quoted compression":
				request.Configuration[".yarnrc.yml"] = []byte(setting)
			case strings.HasPrefix(setting, "supportedArchitectures:"):
				request.Configuration[".yarnrc.yml"] = []byte(strings.Split(baseRC(), "supportedArchitectures:\n")[0] + setting)
			default:
				request.Configuration[".yarnrc.yml"] = []byte(baseRC() + setting)
			}
			_, err := Parse(request)
			if ErrorCode(err) != CodeLockFormatUnsupported {
				t.Fatalf("error = %v (%s), want %s", err, ErrorCode(err), CodeLockFormatUnsupported)
			}
		})
	}
}

func TestN11PluginsGitAndPatchesFailClosed(t *testing.T) {
	t.Run("local plugin", func(t *testing.T) {
		r := baseParseRequest(baseLock())
		r.Configuration[".yarnrc.yml"] = []byte(baseRC() + "plugins:\n  - path: .yarn/plugins/local.cjs\n    spec: local\n")
		_, err := Parse(r)
		if ErrorCode(err) != CodeManagerPluginUndeclared {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("builtin drift", func(t *testing.T) {
		r := baseParseRequest(baseLock())
		r.BuiltinPluginSet = []string{"@yarnpkg/plugin-npm"}
		_, err := Parse(r)
		if ErrorCode(err) != CodeManagerPluginUndeclared {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("git", func(t *testing.T) {
		r := baseParseRequest(strings.Replace(baseLock(), "a@npm:1.0.0", "a@git:deadbeef", 1))
		_, err := Parse(r)
		if ErrorCode(err) != CodeOriginUnpinned {
			t.Fatalf("error = %v (%s)", err, ErrorCode(err))
		}
	})
	t.Run("undeclared patch", func(t *testing.T) {
		r := baseParseRequest(baseLock())
		r.Patches = map[string][]byte{".yarn/patches/a.patch": []byte("diff --git a/a b/a\n")}
		_, err := Parse(r)
		if ErrorCode(err) != CodeManagerPluginUndeclared {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestN11DeclaredPatchBytesAreBoundIntoClosureIdentity(t *testing.T) {
	patchPath := ".yarn/patches/a.patch"
	patchResolution := "a@patch:a@npm%3A1.0.0#./.yarn/patches/a.patch::version=1.0.0&hash=deadbeef"
	patchSelector := "patch:a@npm%3A1.0.0#./.yarn/patches/a.patch"
	lock := "__metadata:\n  version: 8\n  cacheKey: 10c0\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    a: \"" + patchSelector + "\"\n  languageName: unknown\n  linkType: soft\n\"a@" + patchSelector + "\":\n  version: 1.0.0\n  resolution: \"" + patchResolution + "\"\n  checksum: \"" + checksumA + "\"\n  languageName: node\n  linkType: hard\n"
	request := baseParseRequest(lock)
	request.Manifests["package.json"] = []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","dependencies":{"a":"patch:a@npm%3A1.0.0#./.yarn/patches/a.patch"}}`)
	request.Patches = map[string][]byte{patchPath: []byte("diff --git a/index.js b/index.js\n")}
	first, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Patches) != 1 || first.Patches[0].Path != patchPath || first.Patches[0].Locator != patchResolution {
		t.Fatalf("patch evidence = %+v", first.Patches)
	}
	request.Patches[patchPath] = []byte("diff --git a/index.js b/index.js\n+changed\n")
	second, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	if first.LockDigest == second.LockDigest || first.Patches[0].SHA256 == second.Patches[0].SHA256 {
		t.Fatalf("patch drift did not change closure identity: %s", first.LockDigest)
	}
}

func TestN04N05ArchiveNormalizationLifecycleAndNativeGates(t *testing.T) {
	payload := testTGZ(t, map[string][]byte{"package.json": []byte(`{"name":"a","version":"1.0.0"}`), "index.js": []byte("module.exports = 1\n")})
	pkg := Package{Name: "a", Version: "1.0.0"}
	one, err := normalizeCacheZip(pkg, payload, 0)
	if err != nil {
		t.Fatal(err)
	}
	two, err := normalizeCacheZip(pkg, payload, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(one, two) {
		t.Fatal("normalization is not deterministic")
	}
	manifest, inspection, files, err := inspectZip(one)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "a" || inspection.bindingGYP || len(files) != 2 {
		t.Fatalf("zip inspection = %+v %+v %d", manifest, inspection, len(files))
	}
	lifecycle := packageManifest{Name: "a", Version: "1.0.0", Scripts: map[string]string{"postinstall": "curl example.test"}}
	if code := ErrorCode(reconcileEmbeddedMetadata(Package{Key: "a", Name: "a", Version: "1.0.0", Dependencies: map[string]string{}, OptionalDependencies: map[string]string{}}, lifecycle, tarInspection{})); code != CodeHookUndeclared {
		t.Fatalf("lifecycle code = %s", code)
	}
	native := packageManifest{Name: "a", Version: "1.0.0", Scripts: map[string]string{}}
	if code := ErrorCode(reconcileEmbeddedMetadata(Package{Key: "a", Name: "a", Version: "1.0.0", Dependencies: map[string]string{}, OptionalDependencies: map[string]string{}}, native, tarInspection{bindingGYP: true})); code != CodeNativeBuildUnsupported {
		t.Fatalf("native code = %s", code)
	}
}

func TestExternalPeerMetadataCannotWidenLockAuthority(t *testing.T) {
	packageFiles := map[string][]byte{
		"package.json": []byte(`{"name":"is-number","version":"7.0.0","peerDependencies":{"react":"^18.0.0"},"peerDependenciesMeta":{"react":{"optional":true}}}`),
		"index.js":     []byte("module.exports = true\n"),
	}
	t.Run("exact lock metadata", func(t *testing.T) {
		fixture := newModernExecutionFixtureWithPackageFiles(t, "pnp", packageFiles)
		lockBytes, readErr := os.ReadFile(filepath.Join(fixture.root, "yarn.lock"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		lockBytes = []byte(strings.Replace(string(lockBytes), "  languageName: node\n  linkType: hard\n", "  peerDependencies:\n    react: \"^18.0.0\"\n  peerDependenciesMeta:\n    react:\n      optional: true\n  languageName: node\n  linkType: hard\n", 1))
		mustWrite(t, fixture.root, "yarn.lock", lockBytes)
		manifest, manifestErr := os.ReadFile(filepath.Join(fixture.root, "package.json"))
		if manifestErr != nil {
			t.Fatal(manifestErr)
		}
		rc, rcErr := os.ReadFile(filepath.Join(fixture.root, ".yarnrc.yml"))
		if rcErr != nil {
			t.Fatal(rcErr)
		}
		fixture.graph, readErr = Parse(ParseRequest{LockName: "yarn.lock", LockBytes: lockBytes, Manifests: map[string][]byte{"package.json": manifest}, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}})
		if readErr != nil {
			t.Fatal(readErr)
		}
		capture := captureModernFixture(t, fixture)
		locked := capture.Graph.Packages[capture.Graph.packageIndex["is-number@npm:7.0.0"]]
		if locked.PeerDependencies["react"] != "^18.0.0" || !locked.PeerOptional["react"] {
			t.Fatalf("lock-authoritative peer metadata changed during capture: %+v / %+v", locked.PeerDependencies, locked.PeerOptional)
		}
	})

	fixture := newModernExecutionFixtureWithPackageFiles(t, "pnp", packageFiles)
	capture, err := captureModernFixtureError(t, fixture)
	if ErrorCode(err) != CodeMetadataMismatch {
		t.Fatalf("error = %v (%s), want %s", err, ErrorCode(err), CodeMetadataMismatch)
	}
	if capture != nil {
		t.Fatalf("peer metadata drift emitted a capture and could reach manager/build/publication: %+v", capture.Evidence)
	}
	for _, generated := range []string{".yarn/cache", ".yarn/install-state.gz"} {
		if generated == ".yarn/install-state.gz" {
			payload, readErr := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(generated)))
			if readErr != nil || string(payload) != "poison\n" {
				t.Fatalf("preseeded state changed before rejection: path=%s payload=%q error=%v", generated, payload, readErr)
			}
			continue
		}
		if _, statErr := os.Stat(filepath.Join(fixture.root, filepath.FromSlash(generated))); !os.IsNotExist(statErr) {
			t.Fatalf("manager-derived path exists after pre-execution rejection: %s (%v)", generated, statErr)
		}
	}
}

func TestS02N03TamperedOrMissingArchiveFailsBeforeExecution(t *testing.T) {
	t.Run("tampered digest", func(t *testing.T) {
		fixture := newModernExecutionFixture(t)
		for key, archive := range fixture.archives {
			archive.SHA256 = string(digestID([]byte("substituted")))
			fixture.archives[key] = archive
		}
		_, err := captureModernFixtureError(t, fixture)
		if ErrorCode(err) != CodeIntegrityMismatch {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("missing archive", func(t *testing.T) {
		fixture := newModernExecutionFixture(t)
		for key := range fixture.archives {
			delete(fixture.archives, key)
		}
		_, err := captureModernFixtureError(t, fixture)
		if ErrorCode(err) != CodeOfflineInputMissing {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("cache identity", func(t *testing.T) {
		fixture := newModernExecutionFixture(t)
		for key, archive := range fixture.archives {
			archive.CacheName = "substituted.zip"
			fixture.archives[key] = archive
		}
		_, err := captureModernFixtureError(t, fixture)
		if ErrorCode(err) != CodeIntegrityMismatch {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestS04N08LifecycleAndUndeclaredGeneratorPluginFailBeforeExecution(t *testing.T) {
	tests := map[string]map[string][]byte{
		"networking lifecycle": {
			"package.json": []byte(`{"name":"is-number","version":"7.0.0","scripts":{"postinstall":"curl https://example.test"}}`),
			"index.js":     []byte("module.exports = true\n"),
		},
		"generator plugin": {
			"package.json": []byte(`{"name":"is-number","version":"7.0.0","scripts":{"prepare":"node generator.js --plugin undeclared"}}`),
			"generator.js": []byte("require('undeclared-plugin')\n"),
		},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newModernExecutionFixtureWithPackageFiles(t, "pnp", files)
			_, err := captureModernFixtureError(t, fixture)
			if ErrorCode(err) != CodeHookUndeclared {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestS05N06CompiledPayloadDirectRenamedAndNestedFailsBeforeExecution(t *testing.T) {
	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	tests := map[string]map[string][]byte{
		"direct": {
			"package.json": []byte(`{"name":"is-number","version":"7.0.0"}`),
			"addon.wasm":   wasm,
		},
		"renamed": {
			"package.json":  []byte(`{"name":"is-number","version":"7.0.0"}`),
			"looks-safe.js": wasm,
		},
		"nested": {
			"package.json":      []byte(`{"name":"is-number","version":"7.0.0"}`),
			"assets/nested.zip": testZip(t, map[string][]byte{"module.bin": wasm}),
		},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newModernExecutionFixtureWithPackageFiles(t, "pnp", files)
			_, err := captureModernFixtureError(t, fixture)
			if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
				t.Fatalf("error = %v (%s)", err, artifactpolicy.ErrorCode(err))
			}
		})
	}
}

func TestS07MissingTransitiveLockEdgeFailsClosed(t *testing.T) {
	lock := strings.Replace(baseLock(), "  dependencies:\n    b: \"npm:1.0.0\"\n", "  dependencies:\n    missing: \"npm:1.0.0\"\n", 1)
	_, err := Parse(baseParseRequest(lock))
	if ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("error = %v", err)
	}
}

func TestN07ShippedGeneratedTextAndSourceMapRemainAdmittedSource(t *testing.T) {
	fixture := newModernExecutionFixtureWithPackageFiles(t, "pnp", map[string][]byte{
		"package.json":      []byte(`{"name":"is-number","version":"7.0.0"}`),
		"dist/index.js":     []byte("module.exports=()=>42//# sourceMappingURL=index.js.map\n"),
		"dist/index.js.map": []byte(`{"version":3,"sources":["../src/index.ts"],"names":[],"mappings":"AAAA"}`),
	})
	capture := captureModernFixture(t, fixture)
	if len(capture.Evidence.Tarballs) != 1 || !capture.Evidence.Tarballs[0].ArtifactManifestID.Valid() {
		t.Fatalf("capture evidence = %+v", capture.Evidence)
	}
}

func TestN09WorkspaceEscapeAndPostCheckpointDriftFailClosed(t *testing.T) {
	t.Run("workspace escape", func(t *testing.T) {
		request := baseParseRequest(baseLock())
		request.Manifests["../outside/package.json"] = []byte(`{"name":"outside","version":"1.0.0"}`)
		_, err := Parse(request)
		if ErrorCode(err) != CodeLocalPathEscape {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("manifest drift", func(t *testing.T) {
		fixture := newModernExecutionFixture(t)
		mustWrite(t, fixture.root, "package.json", []byte(`{"name":"fixture","version":"1.0.1","private":true,"packageManager":"yarn@4.9.2","dependencies":{"is-number":"npm:7.0.0"}}`))
		_, err := captureModernFixtureError(t, fixture)
		if ErrorCode(err) != CodeLockStale {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestN13PreseededStateDiscardedPatchPreservedAndPluginsRejected(t *testing.T) {
	source := t.TempDir()
	mustWrite(t, source, "package.json", []byte(`{"name":"root","version":"1.0.0"}`))
	mustWrite(t, source, "yarn.lock", []byte("lock\n"))
	mustWrite(t, source, ".yarnrc.yml", []byte(baseRC()))
	mustWrite(t, source, ".pnp.cjs", []byte("poison"))
	mustWrite(t, source, ".yarn/install-state.gz", []byte("poison"))
	mustWrite(t, source, ".yarn/cache/poison.zip", []byte("poison"))
	mustWrite(t, source, ".yarn/patches/a.patch", []byte("patch\n"))
	destination := t.TempDir()
	discarded, err := copyProjectSource(source, destination, "")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(discarded, "\n")
	for _, want := range []string{".pnp.cjs", ".yarn/cache", ".yarn/install-state.gz"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("discarded = %v, missing %s", discarded, want)
		}
	}
	if _, err = os.Stat(filepath.Join(destination, ".yarn/patches/a.patch")); err != nil {
		t.Fatalf("patch not preserved: %v", err)
	}
	pluginSource := t.TempDir()
	mustWrite(t, pluginSource, "package.json", []byte(`{"name":"root","version":"1.0.0"}`))
	mustWrite(t, pluginSource, "yarn.lock", []byte("lock\n"))
	mustWrite(t, pluginSource, ".yarnrc.yml", []byte(baseRC()))
	mustWrite(t, pluginSource, ".yarn/plugins/local.cjs", []byte("module.exports = {}\n"))
	_, err = copyProjectSource(pluginSource, t.TempDir(), "")
	if ErrorCode(err) != CodeManagerPluginUndeclared {
		t.Fatalf("plugin error = %v", err)
	}
}

func TestN11CapturedUndeclaredPatchFailsBeforeManagerExecution(t *testing.T) {
	fixture := newModernExecutionFixture(t)
	mustWrite(t, fixture.root, ".yarn/patches/undeclared.patch", []byte("diff --git a/index.js b/index.js\n"))
	capture, err := captureModernFixtureError(t, fixture)
	if capture != nil || ErrorCode(err) != CodeManagerPluginUndeclared {
		t.Fatalf("capture = %v, error = %v", capture, err)
	}
}

func TestS03S08N12N13ProtectedPrivateCachePnPReplayAndInvoke(t *testing.T) {
	fixture := newModernExecutionFixture(t)
	capture := captureModernFixture(t, fixture)
	if capture.Evidence.ProfileID != ProfileID || len(capture.Evidence.Tarballs) != 1 || !capture.Evidence.NodeCaptureGraphID.Valid() {
		t.Fatalf("capture evidence = %+v", capture.Evidence)
	}
	runner := &modernFakeRunner{capture: capture}
	runner.authority = makeModernExecutionContext(t, capture, runner)
	cacheRoot := filepath.Join(t.TempDir(), "private-cache")
	cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), runner.authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
	if cache.Receipt.SchemaID != "yarn-modern-private-cache-receipt-v1" || len(cache.Receipt.Files) != 1 {
		t.Fatalf("cache receipt = %+v", cache.Receipt)
	}
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, runner.authority)
	if err != nil {
		t.Fatal(err)
	}
	if runner.starts != 1 || !containsStrings(runner.install.Argv, "install", "--immutable", "--immutable-cache", "--mode=skip-build") {
		t.Fatalf("install permit = %+v starts=%d", runner.install, runner.starts)
	}
	if runner.install.Network != "none" || runner.install.Environment["YARN_ENABLE_NETWORK"] != "0" || runner.install.Environment["YARN_ENABLE_SCRIPTS"] != "0" {
		t.Fatalf("offline environment = %+v", runner.install)
	}
	if _, err = os.Stat(filepath.Join(materialized.Root, ".pnp.cjs")); err != nil {
		t.Fatalf("PnP loader not regenerated: %v", err)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", []string{"--fixture"}, runner.authority); err != nil {
		t.Fatal(err)
	}
	if runner.starts != 2 || !equalStrings(runner.invoke.Argv, []string{"--require", "./.pnp.cjs", "index.js", "--fixture"}) {
		t.Fatalf("PnP invocation = %+v starts=%d", runner.invoke, runner.starts)
	}
	if err = makeTestTreeWritable(cache.root); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(cache.root, ".yarn", "cache", cache.Receipt.Files[0].Path)); err != nil {
		t.Fatal(err)
	}
	_, err = Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "missing"), WorkRoot: t.TempDir()}, runner.authority)
	if ErrorCode(err) != CodeOfflineInputMissing {
		t.Fatalf("missing cache error = %v", err)
	}
}

func TestN13NonfunctionalPnPLoaderCannotMaterializeOrPublish(t *testing.T) {
	fixture := newModernExecutionFixture(t)
	capture := captureModernFixture(t, fixture)
	runner := &modernFakeRunner{capture: capture, invalidPnP: true}
	runner.authority = makeModernExecutionContext(t, capture, runner)
	cacheRoot := filepath.Join(t.TempDir(), "private-cache")
	cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), runner.authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
	destination := filepath.Join(t.TempDir(), "materialized")
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: destination, WorkRoot: t.TempDir()}, runner.authority)
	if materialized != nil || ErrorCode(err) != CodeMetadataMismatch || runner.starts != 1 {
		t.Fatalf("materialized = %v, error = %v, starts = %d", materialized, err, runner.starts)
	}
	if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
		t.Fatalf("failed PnP state was published: %v", statErr)
	}
}

func TestN01RealPinnedYarnPnPInvokeThroughVerifiedExecutor(t *testing.T) {
	yarnJS := os.Getenv("CURATOR_TEST_YARN_MODERN_JS")
	if yarnJS == "" {
		t.Skip("set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool")
	}
	fixture := newModernExecutionFixture(t)
	capture := captureModernFixture(t, fixture)
	runner := newModernConcreteRunner(t, yarnJS)
	provider := newModernVerifiedProvider(runner)
	authority := makeModernVerifiedExecutionContext(t, capture, runner, provider)
	cacheRoot := filepath.Join(t.TempDir(), "private-cache")
	cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, authority)
	if err != nil {
		t.Fatalf("real Yarn materialization failed: %v", err)
	}
	receipt, err := Invoke(t.Context(), materialized, "index.js", nil, authority)
	if err != nil {
		t.Fatalf("real protected PnP dependency invocation failed: %v", err)
	}
	if receipt.Audit.Network != "none" || provider.starts != 2 || len(runner.launches) != 2 {
		t.Fatalf("verified execution evidence = receipt=%+v starts=%d launches=%+v", receipt.Audit, provider.starts, runner.launches)
	}
	if !equalPrefix(runner.launches[1].Argv, []string{"--require", "./.pnp.cjs", "index.js"}) {
		t.Fatalf("real Node invocation did not preload regenerated PnP state: %+v", runner.launches[1])
	}
}

func TestN10RealPinnedYarnWorkspacePeerPnPInvokeThroughVerifiedExecutor(t *testing.T) {
	yarnJS := os.Getenv("CURATOR_TEST_YARN_MODERN_JS")
	if yarnJS == "" {
		t.Skip("set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool")
	}
	fixture := newModernPeerWorkspaceFixture(t)
	capture := captureModernFixture(t, fixture)
	runner := newModernConcreteRunner(t, yarnJS)
	provider := newModernVerifiedProvider(runner)
	authority := makeModernVerifiedExecutionContext(t, capture, runner, provider)
	cacheRoot := filepath.Join(t.TempDir(), "private-peer-cache")
	cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized-peer"), WorkRoot: t.TempDir()}, authority)
	if err != nil {
		t.Fatalf("real Yarn peer materialization failed: %v", err)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, authority); err != nil {
		t.Fatalf("real protected peer PnP invocation failed: %v", err)
	}
	if provider.starts != 2 || len(runner.launches) != 2 || !equalPrefix(runner.launches[1].Argv, []string{"--require", "./.pnp.cjs", "index.js"}) {
		t.Fatalf("peer execution launches = starts:%d launches:%+v", provider.starts, runner.launches)
	}
}

func TestN10RealPinnedYarnRemotePeerContextsThroughVerifiedExecutor(t *testing.T) {
	yarnJS := os.Getenv("CURATOR_TEST_YARN_MODERN_JS")
	if yarnJS == "" {
		t.Skip("set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool")
	}
	fixture := newModernTwoRemotePeerContextsFixture(t)
	capture := captureModernFixture(t, fixture)
	runner := newModernConcreteRunner(t, yarnJS)
	provider := newModernVerifiedProvider(runner)
	authority := makeModernVerifiedExecutionContext(t, capture, runner, provider)
	cacheRoot := filepath.Join(t.TempDir(), "private-remote-peer-cache")
	cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized-remote-peers"), WorkRoot: t.TempDir()}, authority)
	if err != nil {
		t.Fatalf("real Yarn remote peer materialization failed: %v", err)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, authority); err != nil {
		t.Fatalf("real remote peer invocation failed: %v", err)
	}
}

func TestN01NodeModulesLinkerReconcilesInstalledPayload(t *testing.T) {
	fixture := newModernExecutionFixtureForLinker(t, "node-modules")
	capture := captureModernFixture(t, fixture)
	runner := &modernFakeRunner{capture: capture}
	runner.authority = makeModernExecutionContext(t, capture, runner)
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	cache, err := BuildPrivateCache(t.Context(), capture, cacheRoot, t.TempDir(), runner.authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(cacheRoot) })
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, runner.authority)
	if err != nil {
		t.Fatal(err)
	}
	if !containsStrings(materialized.MaterializedPackages, "node_modules/is-number") {
		t.Fatalf("packages = %v", materialized.MaterializedPackages)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, runner.authority); err != nil {
		t.Fatal(err)
	}
	if !equalStrings(runner.invoke.Argv, []string{"index.js"}) {
		t.Fatalf("node-modules invocation = %+v", runner.invoke)
	}
}

type modernExecutionFixture struct {
	root, work string
	graph      Graph
	archives   map[string]RawArchive
}

func newModernExecutionFixture(t *testing.T) modernExecutionFixture {
	return newModernExecutionFixtureForLinker(t, "pnp")
}
func newModernExecutionFixtureForLinker(t *testing.T, linker string) modernExecutionFixture {
	return newModernExecutionFixtureWithPackageFiles(t, linker, map[string][]byte{
		"package.json": []byte(`{"name":"is-number","version":"7.0.0"}`),
		"index.js":     []byte("module.exports = true\n"),
	})
}
func newModernExecutionFixtureWithPackageFiles(t *testing.T, linker string, packageFiles map[string][]byte) modernExecutionFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	tgz := testTGZ(t, packageFiles)
	pkg := Package{Name: "is-number", Version: "7.0.0"}
	zipBytes, err := normalizeCacheZip(pkg, tgz, 0)
	if err != nil {
		t.Fatal(err)
	}
	checksum := CacheChecksum(zipBytes, "10c0")
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n\"fixture@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"fixture@workspace:.\"\n  dependencies:\n    is-number: \"npm:7.0.0\"\n  languageName: unknown\n  linkType: soft\n\n\"is-number@npm:7.0.0\":\n  version: 7.0.0\n  resolution: \"is-number@npm:7.0.0\"\n  checksum: " + checksum + "\n  languageName: node\n  linkType: hard\n"
	manifest := []byte(`{"name":"fixture","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","dependencies":{"is-number":"npm:7.0.0"}}`)
	rc := []byte(strings.Replace(baseRC(), "nodeLinker: pnp", "nodeLinker: "+linker, 1))
	mustWrite(t, root, "package.json", manifest)
	mustWrite(t, root, "yarn.lock", []byte(lock))
	mustWrite(t, root, ".yarnrc.yml", rc)
	mustWrite(t, root, "index.js", []byte("require('is-number')\n"))
	mustWrite(t, root, ".pnp.cjs", []byte("poison\n"))
	mustWrite(t, root, ".yarn/install-state.gz", []byte("poison\n"))
	request := ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": manifest}, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}}
	graph, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(t.TempDir(), "is-number.zip")
	if err = os.WriteFile(archivePath, zipBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	archives := map[string]RawArchive{}
	for _, candidate := range graph.Packages {
		if candidate.BaseKey == "" && candidate.Resolution != "" {
			cacheName, nameErr := yarnCacheName(candidate)
			if nameErr != nil {
				t.Fatal(nameErr)
			}
			archives[candidate.Key] = RawArchive{Path: archivePath, Format: "zip", SHA256: string(digestID(zipBytes)), YarnChecksum: checksum, CacheName: cacheName}
		}
	}
	return modernExecutionFixture{root: root, work: t.TempDir(), graph: graph, archives: archives}
}

func newModernWorkspaceCycleFixture(t *testing.T) modernExecutionFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace-cycle-project")
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n\"a@workspace:*, a@workspace:packages/a\":\n  version: 0.0.0-use.local\n  resolution: \"a@workspace:packages/a\"\n  dependencies:\n    b: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n\n\"b@workspace:*, b@workspace:packages/b\":\n  version: 0.0.0-use.local\n  resolution: \"b@workspace:packages/b\"\n  dependencies:\n    a: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    a: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n"
	manifests := map[string][]byte{
		"package.json":            []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","workspaces":["packages/*"],"dependencies":{"a":"workspace:*"}}`),
		"packages/a/package.json": []byte(`{"name":"a","version":"1.0.0","dependencies":{"b":"workspace:*"}}`),
		"packages/b/package.json": []byte(`{"name":"b","version":"1.0.0","dependencies":{"a":"workspace:*"}}`),
	}
	rc := []byte(baseRC())
	files := map[string][]byte{
		"yarn.lock": []byte(lock), ".yarnrc.yml": rc,
		"index.js":            []byte("require('a')\n"),
		"packages/a/index.js": []byte("module.exports = require('b')\n"),
		"packages/b/index.js": []byte("module.exports = require('a')\n"),
		".pnp.cjs":            []byte("poison\n"), ".yarn/install-state.gz": []byte("poison\n"),
	}
	for name, payload := range manifests {
		files[name] = payload
	}
	for name, payload := range files {
		mustWrite(t, root, name, payload)
	}
	graph, err := Parse(ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: manifests, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}})
	if err != nil {
		t.Fatal(err)
	}
	return modernExecutionFixture{root: root, work: t.TempDir(), graph: graph, archives: map[string]RawArchive{}}
}

func newModernWorkspacePeerCycleFixture(t *testing.T) modernExecutionFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace-peer-cycle-project")
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n\"a@workspace:*, a@workspace:packages/a\":\n  version: 0.0.0-use.local\n  resolution: \"a@workspace:packages/a\"\n  dependencies:\n    b: \"workspace:*\"\n    host: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n\n\"b@workspace:*, b@workspace:packages/b\":\n  version: 0.0.0-use.local\n  resolution: \"b@workspace:packages/b\"\n  dependencies:\n    a: \"workspace:*\"\n  peerDependencies:\n    host: \"*\"\n  languageName: unknown\n  linkType: soft\n\n\"host@workspace:*, host@workspace:packages/host\":\n  version: 0.0.0-use.local\n  resolution: \"host@workspace:packages/host\"\n  languageName: unknown\n  linkType: soft\n\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    a: \"workspace:*\"\n    host: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n"
	manifests := map[string][]byte{
		"package.json":               []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","workspaces":["packages/*"],"dependencies":{"a":"workspace:*","host":"workspace:*"}}`),
		"packages/a/package.json":    []byte(`{"name":"a","version":"1.0.0","dependencies":{"b":"workspace:*","host":"workspace:*"}}`),
		"packages/b/package.json":    []byte(`{"name":"b","version":"1.0.0","dependencies":{"a":"workspace:*"},"peerDependencies":{"host":"*"}}`),
		"packages/host/package.json": []byte(`{"name":"host","version":"1.0.0"}`),
	}
	rc := []byte(baseRC())
	files := map[string][]byte{
		"yarn.lock": []byte(lock), ".yarnrc.yml": rc,
		"index.js":               []byte("require('a')\n"),
		"packages/a/index.js":    []byte("module.exports = require('b')\n"),
		"packages/b/index.js":    []byte("module.exports = require('a')\n"),
		"packages/host/index.js": []byte("module.exports = 'host'\n"),
		".pnp.cjs":               []byte("poison\n"), ".yarn/install-state.gz": []byte("poison\n"),
	}
	for name, payload := range manifests {
		files[name] = payload
	}
	for name, payload := range files {
		mustWrite(t, root, name, payload)
	}
	graph, err := Parse(ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: manifests, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}})
	if err != nil {
		t.Fatal(err)
	}
	return modernExecutionFixture{root: root, work: t.TempDir(), graph: graph, archives: map[string]RawArchive{}}
}

func newModernRemoteCycleFixture(t *testing.T) modernExecutionFixture {
	t.Helper()
	type remotePackage struct {
		name, dependency, source string
	}
	packages := []remotePackage{
		{name: "a", dependency: "b", source: "module.exports = require('b')\n"},
		{name: "b", dependency: "a", source: "module.exports = require('a')\n"},
	}
	type archiveRecord struct{ checksum, path, cacheName string }
	records := map[string]archiveRecord{}
	for _, pkg := range packages {
		manifest := []byte(`{"name":"` + pkg.name + `","version":"1.0.0","dependencies":{"` + pkg.dependency + `":"npm:1.0.0"}}`)
		tgz := testTGZ(t, map[string][]byte{"package.json": manifest, "index.js": []byte(pkg.source)})
		identity := Package{Name: pkg.name, Version: "1.0.0", Resolution: pkg.name + "@npm:1.0.0", Key: pkg.name + "@npm:1.0.0"}
		zipBytes, err := normalizeCacheZip(identity, tgz, 0)
		if err != nil {
			t.Fatal(err)
		}
		checksum := CacheChecksum(zipBytes, "10c0")
		identity.Checksum = checksum
		cacheName, err := yarnCacheName(identity)
		if err != nil {
			t.Fatal(err)
		}
		archivePath := filepath.Join(t.TempDir(), pkg.name+"-1.0.0.zip")
		if err = os.WriteFile(archivePath, zipBytes, 0o600); err != nil {
			t.Fatal(err)
		}
		records[identity.Key] = archiveRecord{checksum: checksum, path: archivePath, cacheName: cacheName}
	}
	entry := func(name, dependency string) string {
		record := records[name+"@npm:1.0.0"]
		return "\"" + name + "@npm:1.0.0\":\n  version: 1.0.0\n  resolution: \"" + name + "@npm:1.0.0\"\n  dependencies:\n    " + dependency + ": \"npm:1.0.0\"\n  checksum: " + record.checksum + "\n  languageName: node\n  linkType: hard\n"
	}
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n" + entry("a", "b") + "\n" + entry("b", "a") + "\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    a: \"npm:1.0.0\"\n  languageName: unknown\n  linkType: soft\n"
	manifest := []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","dependencies":{"a":"npm:1.0.0"}}`)
	rc := []byte(baseRC())
	root := filepath.Join(t.TempDir(), "remote-cycle-project")
	for name, payload := range map[string][]byte{"package.json": manifest, "yarn.lock": []byte(lock), ".yarnrc.yml": rc, "index.js": []byte("require('a')\n"), ".pnp.cjs": []byte("poison\n"), ".yarn/install-state.gz": []byte("poison\n")} {
		mustWrite(t, root, name, payload)
	}
	graph, err := Parse(ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": manifest}, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}})
	if err != nil {
		t.Fatal(err)
	}
	archives := map[string]RawArchive{}
	for key, record := range records {
		payload, err := os.ReadFile(record.path)
		if err != nil {
			t.Fatal(err)
		}
		archives[key] = RawArchive{Path: record.path, Format: "zip", SHA256: string(digestID(payload)), YarnChecksum: record.checksum, CacheName: record.cacheName}
	}
	return modernExecutionFixture{root: root, work: t.TempDir(), graph: graph, archives: archives}
}

func newModernPeerWorkspaceFixture(t *testing.T) modernExecutionFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "peer-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n\"host@workspace:*, host@workspace:packages/host\":\n  version: 0.0.0-use.local\n  resolution: \"host@workspace:packages/host\"\n  languageName: unknown\n  linkType: soft\n\n\"plugin@workspace:*, plugin@workspace:packages/plugin\":\n  version: 0.0.0-use.local\n  resolution: \"plugin@workspace:packages/plugin\"\n  peerDependencies:\n    host: \"*\"\n  languageName: unknown\n  linkType: soft\n\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    host: \"workspace:*\"\n    plugin: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n"
	rootManifest := []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","workspaces":["packages/*"],"dependencies":{"host":"workspace:*","plugin":"workspace:*"}}`)
	hostManifest := []byte(`{"name":"host","version":"1.0.0"}`)
	pluginManifest := []byte(`{"name":"plugin","version":"1.0.0","peerDependencies":{"host":"*"}}`)
	rc := []byte(baseRC())
	for name, payload := range map[string][]byte{
		"package.json": rootManifest, "packages/host/package.json": hostManifest, "packages/plugin/package.json": pluginManifest,
		"yarn.lock": []byte(lock), ".yarnrc.yml": rc, "index.js": []byte("require('plugin')\n"),
		"packages/host/index.js": []byte("module.exports = 'host'\n"), "packages/plugin/index.js": []byte("module.exports = require('host')\n"),
		".pnp.cjs": []byte("poison\n"), ".yarn/install-state.gz": []byte("poison\n"),
	} {
		mustWrite(t, root, name, payload)
	}
	graph, err := Parse(ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": rootManifest, "packages/host/package.json": hostManifest, "packages/plugin/package.json": pluginManifest}, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}})
	if err != nil {
		t.Fatal(err)
	}
	return modernExecutionFixture{root: root, work: t.TempDir(), graph: graph, archives: map[string]RawArchive{}}
}

func newModernTwoRemotePeerContextsFixture(t *testing.T) modernExecutionFixture {
	t.Helper()
	type remotePackage struct {
		name, version, dependencies, peers, source string
	}
	packages := []remotePackage{
		{name: "host", version: "1.0.0", source: "module.exports = 'host-one'\n"},
		{name: "host", version: "2.0.0", source: "module.exports = 'host-two'\n"},
		{name: "plugin", version: "1.0.0", peers: `,"peerDependencies":{"host":"*"}`, source: "module.exports = require('host')\n"},
		{name: "consumer-one", version: "1.0.0", dependencies: `,"dependencies":{"host":"npm:1.0.0","plugin":"npm:1.0.0"}`, source: "module.exports = require('plugin')\n"},
		{name: "consumer-two", version: "1.0.0", dependencies: `,"dependencies":{"host":"npm:2.0.0","plugin":"npm:1.0.0"}`, source: "module.exports = require('plugin')\n"},
	}
	type archiveRecord struct {
		checksum, path, cacheName string
	}
	records := map[string]archiveRecord{}
	for _, pkg := range packages {
		manifest := []byte(`{"name":"` + pkg.name + `","version":"` + pkg.version + `"` + pkg.dependencies + pkg.peers + `}`)
		tgz := testTGZ(t, map[string][]byte{"package.json": manifest, "index.js": []byte(pkg.source)})
		identity := Package{Name: pkg.name, Version: pkg.version, Resolution: pkg.name + "@npm:" + pkg.version, Key: pkg.name + "@npm:" + pkg.version}
		zipBytes, err := normalizeCacheZip(identity, tgz, 0)
		if err != nil {
			t.Fatal(err)
		}
		checksum := CacheChecksum(zipBytes, "10c0")
		identity.Checksum = checksum
		cacheName, err := yarnCacheName(identity)
		if err != nil {
			t.Fatal(err)
		}
		archivePath := filepath.Join(t.TempDir(), strings.ReplaceAll(pkg.name+"-"+pkg.version, "/", "-")+".zip")
		if err = os.WriteFile(archivePath, zipBytes, 0o600); err != nil {
			t.Fatal(err)
		}
		records[identity.Key] = archiveRecord{checksum: checksum, path: archivePath, cacheName: cacheName}
	}
	entry := func(key string, dependencies, peers string) string {
		record := records[key]
		_, version, _ := strings.Cut(key, "@npm:")
		return "\"" + key + "\":\n  version: " + version + "\n  resolution: \"" + key + "\"" + dependencies + peers + "\n  checksum: " + record.checksum + "\n  languageName: node\n  linkType: hard\n"
	}
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n"
	lock += entry("consumer-one@npm:1.0.0", "\n  dependencies:\n    host: \"npm:1.0.0\"\n    plugin: \"npm:1.0.0\"", "") + "\n"
	lock += entry("consumer-two@npm:1.0.0", "\n  dependencies:\n    host: \"npm:2.0.0\"\n    plugin: \"npm:1.0.0\"", "") + "\n"
	lock += entry("host@npm:1.0.0", "", "") + "\n" + entry("host@npm:2.0.0", "", "") + "\n"
	lock += entry("plugin@npm:1.0.0", "", "\n  peerDependencies:\n    host: \"*\"") + "\n"
	lock += "\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    consumer-one: \"npm:1.0.0\"\n    consumer-two: \"npm:1.0.0\"\n  languageName: unknown\n  linkType: soft\n"
	rootManifest := []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","dependencies":{"consumer-one":"npm:1.0.0","consumer-two":"npm:1.0.0"}}`)
	rc := []byte(baseRC())
	root := filepath.Join(t.TempDir(), "remote-peer-project")
	for name, payload := range map[string][]byte{"package.json": rootManifest, "yarn.lock": []byte(lock), ".yarnrc.yml": rc, "index.js": []byte("require('consumer-one'); require('consumer-two')\n"), ".pnp.cjs": []byte("poison\n"), ".yarn/install-state.gz": []byte("poison\n")} {
		mustWrite(t, root, name, payload)
	}
	graph, err := Parse(ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": rootManifest}, Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}})
	if err != nil {
		t.Fatal(err)
	}
	archives := map[string]RawArchive{}
	for key, record := range records {
		payload, err := os.ReadFile(record.path)
		if err != nil {
			t.Fatal(err)
		}
		archives[key] = RawArchive{Path: record.path, Format: "zip", SHA256: string(digestID(payload)), YarnChecksum: record.checksum, CacheName: record.cacheName}
	}
	return modernExecutionFixture{root: root, work: t.TempDir(), graph: graph, archives: archives}
}
func captureModernFixture(t *testing.T, fixture modernExecutionFixture) *Capture {
	t.Helper()
	capture, err := captureModernFixtureError(t, fixture)
	if err != nil {
		t.Fatal(err)
	}
	return capture
}
func captureModernFixtureError(t *testing.T, fixture modernExecutionFixture) (*Capture, error) {
	t.Helper()
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := closureexec.NewCaptureStore(storeRoot)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = makeTestTreeWritable(storeRoot) })
	capture, err := CaptureAndAdmit(t.Context(), CaptureRequest{Graph: fixture.graph, ProjectRoot: fixture.root, Archives: fixture.archives, WorkRoot: fixture.work, Store: store, Policy: artifactpolicy.NewService(), PreviousCausalHead: "test-head"})
	return capture, err
}

type modernFakeRunner struct {
	capture    *Capture
	authority  *ExecutionContext
	starts     int
	install    closureexec.DerivationPermit
	invoke     closureexec.DerivationPermit
	invalidPnP bool
}

type modernTestRunner interface {
	closureexec.PortableProcessRunner
	RecheckTool(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
}

type modernConcreteRunner struct {
	executionRoot, sandboxPath        string
	nodePath, yarnPath                string
	nodeRelative, yarnRelative        string
	nodeDigest, yarnDigest, managerID closuregraph.ID
	launches                          []closureexec.ProcessLaunch
}

func newModernConcreteRunner(t *testing.T, yarnJS string) *modernConcreteRunner {
	t.Helper()
	sandboxPath, err := exec.LookPath("sandbox-exec")
	if err != nil {
		t.Skip("OS-level network-denial harness is unavailable")
	}
	yarnJS, err = filepath.Abs(yarnJS)
	if err != nil {
		t.Fatal(err)
	}
	packageRoot := filepath.Dir(filepath.Dir(yarnJS))
	packageBytes, err := os.ReadFile(filepath.Join(packageRoot, "package.json")) // #nosec G304 -- explicitly selected integration tool.
	if err != nil {
		t.Fatal(err)
	}
	var yarnPackage struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(packageBytes, &yarnPackage) != nil || yarnPackage.Version != SupportedYarnVersion {
		t.Fatalf("integration Yarn release = %q, want %s", yarnPackage.Version, SupportedYarnVersion)
	}
	nodeCommand, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	nodePath, err := filepath.EvalSymlinks(nodeCommand)
	if err != nil {
		t.Fatal(err)
	}
	nodePayload, err := os.ReadFile(nodePath) // #nosec G304 -- exact C0-selected Node executable.
	if err != nil {
		t.Fatal(err)
	}
	yarnPayload, err := os.ReadFile(yarnJS) // #nosec G304 -- exact C0-selected Yarn entry point.
	if err != nil {
		t.Fatal(err)
	}
	executionRoot := filepath.Join(t.TempDir(), "execution")
	for _, logical := range []string{"toolchain/node/bin", "toolchain/yarn/bin", "work", "output"} {
		if err = os.MkdirAll(filepath.Join(executionRoot, filepath.FromSlash(logical)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	stagedNode := filepath.Join(executionRoot, "toolchain", "node", "bin", "node")
	stagedYarn := filepath.Join(executionRoot, "toolchain", "yarn", "bin", "yarn.js")
	if err = os.WriteFile(stagedNode, nodePayload, 0o500); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(stagedYarn, yarnPayload, 0o400); err != nil {
		t.Fatal(err)
	}
	nodeRoot := filepath.Dir(filepath.Dir(nodePath))
	libraries, err := filepath.Glob(filepath.Join(nodeRoot, "lib", "libnode*.dylib"))
	if err != nil {
		t.Fatal(err)
	}
	for _, library := range libraries {
		payload, readErr := os.ReadFile(library) // #nosec G304 -- selected runtime adjacent library.
		if readErr != nil {
			t.Fatal(readErr)
		}
		target := filepath.Join(executionRoot, "toolchain", "node", "lib", filepath.Base(library))
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, payload, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	nodeDigest, yarnDigest := digestID(nodePayload), digestID(yarnPayload)
	managerID, err := closuregraph.DomainID("yarn-modern-interpreted-toolchain-v1", map[string]any{"entrypoint_sha256": string(yarnDigest), "node_sha256": string(nodeDigest), "version": SupportedYarnVersion})
	if err != nil {
		t.Fatal(err)
	}
	return &modernConcreteRunner{executionRoot: executionRoot, sandboxPath: sandboxPath, nodePath: stagedNode, yarnPath: stagedYarn, nodeRelative: "toolchain/node/bin/node", yarnRelative: "toolchain/yarn/bin/yarn.js", nodeDigest: nodeDigest, yarnDigest: yarnDigest, managerID: managerID}
}

func (runner *modernConcreteRunner) Run(ctx context.Context, request closureexec.ExecutionRequest) (closureexec.PortableRunResult, error) {
	if err := prepareModernFakeExecution(request, runner.executionRoot); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	executable := filepath.Join(runner.executionRoot, filepath.FromSlash(request.Permit.Executable))
	cwd := filepath.Join(runner.executionRoot, filepath.FromSlash(request.Permit.CWD))
	profile := "(version 1) (allow default) (deny network*)"
	commandArgs := append([]string{"-p", profile, executable}, request.Permit.Argv...)
	command := exec.CommandContext(ctx, runner.sandboxPath, commandArgs...) // #nosec G204 -- exact sandbox, executable, and permit argv.
	command.Dir = cwd
	keys := sortedKeys(request.Permit.Environment)
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+request.Permit.Environment[key])
	}
	runner.launches = append(runner.launches, closureexec.ProcessLaunch{Executable: executable, CWD: cwd, Argv: append([]string(nil), request.Permit.Argv...), Environment: append([]string(nil), command.Env...)})
	output, err := command.CombinedOutput()
	if err != nil {
		return closureexec.PortableRunResult{}, fmt.Errorf("sandboxed modern Yarn/Node failed: %w: %s", err, output)
	}
	return modernPortableResult(runner.executionRoot, request.Permit)
}

func (runner *modernConcreteRunner) RecheckTool(_ context.Context, tool nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	nodePayload, err := os.ReadFile(runner.nodePath) // #nosec G304 -- exact staged Node executable.
	if err != nil || digestID(nodePayload) != runner.nodeDigest {
		return closureexec.ToolchainIdentity{}, fmt.Errorf("selected Node changed: %w", err)
	}
	fingerprint := runner.nodeDigest
	if tool.Role == "package-manager" {
		yarnPayload, readErr := os.ReadFile(runner.yarnPath) // #nosec G304 -- exact staged Yarn entry point.
		if readErr != nil || digestID(yarnPayload) != runner.yarnDigest {
			return closureexec.ToolchainIdentity{}, fmt.Errorf("selected Yarn changed: %w", readErr)
		}
		fingerprint = runner.managerID
	}
	return closureexec.ToolchainIdentity{Fingerprint: fingerprint, ExecutableSHA256: runner.nodeDigest}, nil
}

type modernVerifiedProvider struct {
	runner   *modernConcreteRunner
	identity closureexec.ProviderIdentity
	receipts map[string]closureexec.ProviderCapabilityReceipt
	starts   int
}

func newModernVerifiedProvider(runner *modernConcreteRunner) *modernVerifiedProvider {
	return &modernVerifiedProvider{runner: runner, identity: closureexec.ProviderIdentity{Contract: closureexec.VerifiedProviderContractID, ProviderID: "yarn-modern-test-provider", Version: "1.0.0", BinarySHA256: digestID([]byte("yarn-modern-test-provider")), TrustEvidence: "test-provider-trust"}, receipts: map[string]closureexec.ProviderCapabilityReceipt{}}
}

func (provider *modernVerifiedProvider) Identity() closureexec.ProviderIdentity {
	return provider.identity
}
func (*modernVerifiedProvider) LosslessObservation() bool { return true }
func (provider *modernVerifiedProvider) config() closureexec.AssuranceConfig {
	return closureexec.AssuranceConfig{Mode: closureexec.AssuranceVerified, ProviderID: provider.identity.ProviderID, ProviderVersion: provider.identity.Version, ProviderBinarySHA256: provider.identity.BinarySHA256, ProviderTrustEvidence: provider.identity.TrustEvidence}
}
func (provider *modernVerifiedProvider) Negotiate(_ context.Context, nonce string) (closureexec.ProviderCapabilityReceipt, error) {
	if receipt, ok := provider.receipts[nonce]; ok {
		return receipt, nil
	}
	now := time.Now().UTC()
	receipt := closureexec.ProviderCapabilityReceipt{Provider: provider.identity, Health: "healthy", Nonce: nonce, ObservedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute), Capabilities: []closureexec.CapabilityEvidence{
		{CapabilityID: "total-network-denial-v1", Status: "established"},
		{CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"},
		{CapabilityID: "exact-executable-allowlisting-v1", Status: "established"},
		{CapabilityID: "private-build-root-only-writes-v1", Status: "established"},
		{CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"},
		{CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"},
	}}
	provider.receipts[nonce] = receipt
	return receipt, nil
}
func (provider *modernVerifiedProvider) EnforceAndObserve(ctx context.Context, request closureexec.ExecutionRequest) (closureexec.Audit, error) {
	provider.starts++
	result, err := provider.runner.Run(ctx, request)
	if err != nil {
		return closureexec.Audit{}, err
	}
	outputs := make([]closureexec.DerivationOutput, len(request.Permit.ExpectedEvidence))
	evidence := make([]string, len(request.Permit.ExpectedEvidence))
	for index, expected := range request.Permit.ExpectedEvidence {
		payload, readErr := os.ReadFile(filepath.Join(result.OutputRoot, filepath.FromSlash(expected.Path))) // #nosec G304 -- exact provider-owned evidence path.
		if readErr != nil {
			return closureexec.Audit{}, readErr
		}
		evidence[index] = expected.Path
		outputs[index] = closureexec.DerivationOutput{Path: expected.Path, SchemaID: expected.SchemaID, ArtifactManifestID: expected.ArtifactManifestID, SHA256: digestID(payload), Size: int64(len(payload))}
	}
	return closureexec.Audit{Executable: request.Permit.Executable, CWD: request.Permit.CWD, Argv: append([]string(nil), request.Permit.Argv...), Environment: cloneStrings(request.Permit.Environment), Processes: append([]string(nil), request.Permit.AllowedProcesses...), Reads: append([]string(nil), request.Permit.ReadRoots...), Writes: append([]string(nil), request.Permit.WriteRoots...), Evidence: evidence, Network: "none", ExitCode: result.ExitCode, Outputs: outputs}, nil
}

func (r *modernFakeRunner) RecheckTool(_ context.Context, tool nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	return closureexec.ToolchainIdentity{Fingerprint: tool.Fingerprint, ExecutableSHA256: tool.ExecutableSHA256}, nil
}
func (r *modernFakeRunner) Run(_ context.Context, request closureexec.ExecutionRequest) (closureexec.PortableRunResult, error) {
	r.starts++
	if err := prepareModernFakeExecution(request, r.authority.ExecutionRoot); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	if containsStrings(request.Permit.Argv, "install") {
		r.install = request.Permit
		cwd := filepath.Join(r.authority.ExecutionRoot, filepath.FromSlash(request.Permit.CWD))
		if r.capture.Graph.Layout.NodeLinker == "pnp" {
			loader, loaderErr := fixturePnPLoader(r.capture)
			if loaderErr != nil {
				return closureexec.PortableRunResult{}, loaderErr
			}
			if r.invalidPnP {
				loader = []byte("module.exports = {};\n")
			}
			if err := os.WriteFile(filepath.Join(cwd, ".pnp.cjs"), loader, 0o600); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		} else {
			for _, pkg := range r.capture.Graph.Packages {
				if !pkg.Selected || pkg.Resolution == "" {
					continue
				}
				if err := extractModernCache(r.capture.tarballs[pkg.Key].cacheBytes, pkg.Name, filepath.Join(cwd, "node_modules", filepath.FromSlash(pkg.Name))); err != nil {
					return closureexec.PortableRunResult{}, err
				}
			}
		}
		if err := os.MkdirAll(filepath.Join(cwd, ".yarn"), 0o700); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		if err := os.WriteFile(filepath.Join(cwd, ".yarn", "install-state.gz"), []byte("derived-state\n"), 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
	} else {
		r.invoke = request.Permit
		if r.capture.Graph.Layout.NodeLinker == "pnp" && !equalPrefix(request.Permit.Argv, []string{"--require", "./.pnp.cjs"}) {
			return closureexec.PortableRunResult{}, fmt.Errorf("MODULE_NOT_FOUND: PnP loader was not preloaded")
		}
	}
	return modernPortableResult(r.authority.ExecutionRoot, request.Permit)
}

func fixturePnPLoader(capture *Capture) ([]byte, error) {
	registry := []any{}
	locators := map[string]string{}
	for _, pkg := range capture.Graph.Packages {
		if !pkg.Selected {
			continue
		}
		locator, err := fixturePnPPackageLocator(pkg)
		if err != nil {
			return nil, err
		}
		locators[pkg.Key] = locator
	}
	for _, pkg := range capture.Graph.Packages {
		if !pkg.Selected {
			continue
		}
		locator := locators[pkg.Key]
		parts := strings.SplitN(locator, "\x00", 2)
		location := "./"
		if pkg.BaseKey != "" {
			base := capture.Graph.Packages[capture.Graph.packageIndex[pkg.BaseKey]]
			baseLocation, locationErr := expectedPnPBaseLocation(base, capture.tarballs)
			if locationErr != nil {
				return nil, locationErr
			}
			component, slugErr := yarnVirtualLocatorSlug(locator)
			if slugErr != nil {
				return nil, slugErr
			}
			depth, suffix := "1", strings.TrimPrefix(baseLocation, "./")
			if strings.HasPrefix(baseLocation, "./.yarn/") {
				depth, suffix = "0", strings.TrimPrefix(baseLocation, "./.yarn/")
			}
			location = "./.yarn/__virtual__/" + component + "/" + depth + "/" + suffix
		} else if pkg.Resolution != "" {
			location = "./.yarn/cache/" + capture.tarballs[packageSourceKey(pkg)].cacheName + "/node_modules/" + pkg.Name + "/"
		} else if pkg.Key != "workspace:." {
			location = "./" + strings.TrimSuffix(pkg.WorkspacePath, "/") + "/"
		}
		dependencies := []any{[]any{pkg.Name, parts[1]}}
		for _, edge := range capture.Graph.Edges {
			if edge.From != pkg.Key {
				continue
			}
			if edge.To == "" && edge.Scope == "peer" && edge.Reason == "optional_peer_unresolved" {
				dependencies = append(dependencies, []any{edge.Name, nil})
				continue
			}
			if edge.Selected {
				target := locators[edge.To]
				targetParts := strings.SplitN(target, "\x00", 2)
				value := any(targetParts[1])
				if targetParts[0] != edge.Name {
					value = []any{targetParts[0], targetParts[1]}
				}
				dependencies = append(dependencies, []any{edge.Name, value})
			}
		}
		information := map[string]any{"packageLocation": location, "packageDependencies": dependencies, "linkType": "HARD"}
		if peers := expectedPnPPeers(pkg, capture.Graph); len(peers) > 0 {
			information["packagePeers"] = peers
		}
		registry = append(registry, []any{parts[0], []any{[]any{parts[1], information}}})
	}
	state := map[string]any{"dependencyTreeRoots": []any{map[string]any{"name": capture.Graph.RootName, "reference": "workspace:."}}, "packageRegistryData": registry}
	raw, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	return append(append([]byte("const RAW_RUNTIME_STATE =\n'"), raw...), []byte("';\n\nfunction $$SETUP_STATE() {}\n")...), nil
}

func fixturePnPPackageLocator(pkg Package) (string, error) {
	if pkg.BaseKey == "" {
		return pnpPackageLocator(pkg)
	}
	base := pkg
	base.Key, base.BaseKey, base.PeerContext = pkg.BaseKey, "", nil
	baseLocator, err := pnpPackageLocator(base)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(baseLocator, "\x00", 2)
	entropy := sha512.Sum512([]byte(pkg.Key))
	return parts[0] + "\x00virtual:" + hex.EncodeToString(entropy[:]) + "#" + parts[1], nil
}

func equalPrefix(values, prefix []string) bool {
	return len(values) >= len(prefix) && equalStrings(values[:len(prefix)], prefix)
}
func extractModernCache(payload []byte, packageName, destination string) error {
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}
	prefix := "node_modules/" + packageName + "/"
	for _, member := range reader.File {
		if member.FileInfo().IsDir() || !strings.HasPrefix(member.Name, prefix) {
			continue
		}
		relative := strings.TrimPrefix(member.Name, prefix)
		target := filepath.Join(destination, filepath.FromSlash(relative))
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		stream, openErr := member.Open()
		if openErr != nil {
			return openErr
		}
		data, readErr := io.ReadAll(stream)
		_ = stream.Close()
		if readErr != nil {
			return readErr
		}
		if err = os.WriteFile(target, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}
func makeModernExecutionContext(t *testing.T, capture *Capture, runner modernTestRunner) *ExecutionContext {
	t.Helper()
	executionRoot := filepath.Join(t.TempDir(), "execution")
	if concrete, ok := runner.(*modernConcreteRunner); ok {
		executionRoot = concrete.executionRoot
	}
	for _, dir := range []string{"bin", "work", "output"} {
		if err := os.MkdirAll(filepath.Join(executionRoot, dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	target := capture.Graph.Target
	platform := closuregraph.TargetPlatformPayload{OS: target.OS, Architecture: target.Architecture, ABI: "node", Libc: target.Libc, MinimumRuntime: "bound-by-node-runtime", SDKID: "none", TargetTriple: target.Architecture + "-" + target.OS, Runtime: "node", LanguageModes: map[string]string{"package_manager": "yarn-modern"}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	selection, err := closuregraph.NewSelectionContext(capture.NodeCapture.Graph.RootNodeIDs, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, target.IncludeDev, map[string]string{"os": target.OS, "architecture": target.Architecture, "cpu": target.Architecture, "libc": target.Libc}, map[string]string{}, []string{"yarn-modern-platform-v1"})
	if err != nil {
		t.Fatal(err)
	}
	tool := func(role, seed, version string) nodesource.ToolIdentity {
		return nodesource.ToolIdentity{Role: role, PolicySelector: role + "-v1", ExecutableRelativePath: "bin/node", VersionOutput: version, PlatformABI: target.OS + "-" + target.Architecture, Fingerprint: digestID([]byte(seed + "-tree")), ExecutableSHA256: digestID([]byte("node-executable")), ExecutionDomain: closuregraph.ExecutionTarget}
	}
	runtime := nodesource.RuntimeBinding{Platform: platform, Node: tool("node-runtime", "node", "v25.6.1"), Manager: tool("package-manager", "yarn", SupportedYarnVersion), TargetNodeIDs: append([]closuregraph.ID(nil), capture.NodeCapture.Graph.RootNodeIDs...)}
	runtime.Node.ReadRoots = []string{"toolchain/node"}
	runtime.Manager.EntrypointRelativePath = "toolchain/yarn/bin/yarn.js"
	runtime.Manager.ReadRoots = []string{"toolchain/node", "toolchain/yarn"}
	if concrete, ok := runner.(*modernConcreteRunner); ok {
		runtime.Node.ExecutableRelativePath = concrete.nodeRelative
		runtime.Node.ExecutableSHA256 = concrete.nodeDigest
		runtime.Node.Fingerprint = concrete.nodeDigest
		runtime.Manager.ExecutableRelativePath = concrete.nodeRelative
		runtime.Manager.EntrypointRelativePath = concrete.yarnRelative
		runtime.Manager.ExecutableSHA256 = concrete.nodeDigest
		runtime.Manager.Fingerprint = concrete.managerID
	}
	c0, err := nodesource.NewC0Checkpoint(capture.NodeCapture, selection, runtime)
	if err != nil {
		t.Fatal(err)
	}
	runtime.C0Checkpoint = &c0
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "yarn-modern-platform-v1", EvaluateFunc: evaluateYarnPlatformCondition}
	_, plan, err := nodesource.Close(capture.NodeCapture, selection, runtime, []closuregraph.ConditionEvaluator{evaluator}, closureexec.PortableExecutionPolicyID)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), runner, nil, "test-head")
	if err != nil {
		t.Fatal(err)
	}
	return &ExecutionContext{Executor: executor, Selection: selection, Runtime: runtime, BuildPlan: plan, Recheck: runner.RecheckTool, ExecutionRoot: executionRoot}
}

func makeModernVerifiedExecutionContext(t *testing.T, capture *Capture, runner *modernConcreteRunner, provider *modernVerifiedProvider) *ExecutionContext {
	t.Helper()
	base := makeModernExecutionContext(t, capture, runner)
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "yarn-modern-platform-v1", EvaluateFunc: evaluateYarnPlatformCondition}
	_, plan, err := nodesource.Close(capture.NodeCapture, base.Selection, base.Runtime, []closuregraph.ConditionEvaluator{evaluator}, closureexec.VerifiedExecutionPolicyID)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := closureexec.NewAssuredExecutor(provider.config(), nil, provider, "test-head")
	if err != nil {
		t.Fatal(err)
	}
	return &ExecutionContext{Executor: executor, Selection: base.Selection, Runtime: base.Runtime, BuildPlan: plan, Recheck: runner.RecheckTool, ExecutionRoot: base.ExecutionRoot}
}
func prepareModernFakeExecution(request closureexec.ExecutionRequest, root string) error {
	inputs := map[closuregraph.ID]closureexec.ReplayInput{}
	for _, input := range request.Inputs {
		inputs[input.ReceiptID] = input
		protected, err := input.ProtectedPath()
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(input.MountPath))
		_ = os.RemoveAll(target)
		if input.IsTree() {
			if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			if err = copyWritableTestTree(protected, target); err != nil {
				return err
			}
		} else {
			payload, readErr := os.ReadFile(protected)
			if readErr != nil {
				return readErr
			}
			if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			if err = os.WriteFile(target, payload, 0o600); err != nil {
				return err
			}
		}
	}
	for _, work := range request.Permit.WorkCopies {
		input := inputs[work.ReceiptID]
		source, err := input.ProtectedPath()
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(work.Path))
		_ = os.RemoveAll(target)
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err = copyWritableTestTree(source, target); err != nil {
			return err
		}
	}
	output := filepath.Join(root, "output")
	_ = os.RemoveAll(output)
	return os.MkdirAll(output, 0o700)
}
func modernPortableResult(executionRoot string, permit closureexec.DerivationPermit) (closureexec.PortableRunResult, error) {
	root := filepath.Join(executionRoot, filepath.FromSlash(permit.Environment["CURATOR_OUTPUT_ROOT"]))
	for _, expected := range permit.ExpectedEvidence {
		target := filepath.Join(root, filepath.FromSlash(expected.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		if err := os.WriteFile(target, nil, 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
	}
	return closureexec.PortableRunResult{ExitCode: 0, OutputRoot: root}, nil
}
func copyWritableTestTree(source, destination string) error {
	if err := os.Mkdir(destination, 0o700); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == source {
			return nil
		}
		rel, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("test input contains link %s", rel)
		}
		if entry.IsDir() {
			return os.Mkdir(target, 0o700)
		}
		payload, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o600)
	})
}
func makeTestTreeWritable(root string) error {
	return filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return os.Chmod(current, 0o700)
		}
		return os.Chmod(current, 0o600)
	})
}
func containsStrings(values []string, wanted ...string) bool {
	for _, want := range wanted {
		found := false
		for _, value := range values {
			found = found || value == want
		}
		if !found {
			return false
		}
	}
	return true
}

func baseParseRequest(lock string) ParseRequest {
	return ParseRequest{LockName: "yarn.lock", LockBytes: []byte(lock), YarnVersion: SupportedYarnVersion, Target: Target{OS: "linux", Architecture: "x64", Libc: "glibc", IncludeDev: true}, Configuration: map[string][]byte{".yarnrc.yml": []byte(baseRC())}, Manifests: map[string][]byte{"package.json": []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","dependencies":{"a":"npm:^1.0.0"}}`)}}
}
func baseRC() string {
	return "nodeLinker: pnp\ncompressionLevel: 0\ncacheFolder: .yarn/cache\nenableGlobalCache: false\nenableNetwork: false\nenableImmutableInstalls: true\nenableScripts: false\nchecksumBehavior: throw\npnpMode: strict\nsupportedArchitectures:\n  os: [linux]\n  cpu: [x64]\n  libc: [glibc]\n"
}
func baseLock() string {
	return "__metadata:\n  version: 8\n  cacheKey: 10c0\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    a: \"npm:^1.0.0\"\n  languageName: unknown\n  linkType: soft\n" + remoteEntries()
}
func workspaceParseRequest() ParseRequest {
	root := []byte(`{"name":"root","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","workspaces":["packages/*"],"dependencies":{"child":"workspace:*"}}`)
	child := []byte(`{"name":"child","version":"2.3.4"}`)
	lock := "__metadata:\n  version: 8\n  cacheKey: 10c0\n\"root@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"root@workspace:.\"\n  dependencies:\n    child: \"workspace:*\"\n  languageName: unknown\n  linkType: soft\n\"child@workspace:*, child@workspace:packages/child\":\n  version: 0.0.0-use.local\n  resolution: \"child@workspace:packages/child\"\n  languageName: unknown\n  linkType: soft\n"
	request := baseParseRequest(lock)
	request.Manifests = map[string][]byte{"package.json": root, "packages/child/package.json": child}
	return request
}
func remoteEntries() string {
	return "\"a@npm:^1.0.0\":\n  version: 1.0.0\n  resolution: \"a@npm:1.0.0\"\n  checksum: \"" + checksumA + "\"\n  dependencies:\n    b: \"npm:1.0.0\"\n  conditions: \"os=linux & cpu=x64\"\n  languageName: node\n  linkType: hard\n\"b@npm:1.0.0\":\n  version: 1.0.0\n  resolution: \"b@npm:1.0.0\"\n  checksum: \"" + checksumB + "\"\n  languageName: node\n  linkType: hard\n"
}
func reverseRemoteEntries() string {
	parts := strings.Split(remoteEntries(), "\"b@npm:1.0.0\":")
	return "\"b@npm:1.0.0\":" + parts[1] + parts[0]
}
func testTGZ(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	gz := gzip.NewWriter(&out)
	tw := tar.NewWriter(gz)
	for _, name := range sortedKeys(files) {
		payload := files[name]
		if err := tw.WriteHeader(&tar.Header{Name: "package/" + name, Mode: 0o644, Size: int64(len(payload))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func testZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for _, name := range sortedKeys(files) {
		header := &zip.FileHeader{Name: name, Method: zip.Store, Modified: time.Unix(0, 0).UTC()}
		header.SetMode(0o644)
		writer, err := zw.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = writer.Write(files[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func mustWrite(t *testing.T, root, relative string, payload []byte) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}
