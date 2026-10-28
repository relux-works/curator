package envunit

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/envregistry"
)

// BackupQuarantine identifies opaque cleanup of an owned backup generation.
const BackupQuarantine = "backup-quarantine"

// BackupAdapter admits one canonical numeric generation below an enumerated
// home, independently of target labels. It grants metadata-only removal, never
// a recursive byte preimage or authority over arbitrary home descendants.
func (u *Unit) BackupAdapter(path string) (string, bool) {
	if u.Action != Maintain && u.Action != Rebind && u.Action != Provision && u.Action != GlobalExtend && u.Action != Takeover {
		return "", false
	}
	match := func(root string) bool {
		base := filepath.Join(root, ".agent-environment-backup")
		n, err := strconv.Atoi(filepath.Base(path))
		return err == nil && n > 0 && strconv.Itoa(n) == filepath.Base(path) && filepath.Dir(path) == base
	}
	for _, h := range u.Homes {
		if match(h.Target) {
			return h.Adapter, true
		}
	}
	for _, n := range u.Native {
		if match(n.Root) {
			return n.Adapter, true
		}
	}
	return "", false
}

// BackupManaged distinguishes the closed managed seed/config namespace from
// the narrower native context/skill backup namespace.
func (u *Unit) BackupManaged(path string) bool {
	for _, h := range u.Homes {
		if filepath.Dir(path) == filepath.Join(h.Target, ".agent-environment-backup") {
			return true
		}
	}
	return false
}

// CheckBackupGeneration inspects only names and metadata in fixed context,
// skill and (for managed homes) seed/config slots. Credential/session slots
// never enter custody. Skill payloads are neither read nor hashed.
func CheckBackupGeneration(path, adapter string, managed bool) error {
	a, err := envregistry.ByID(adapter)
	if err != nil {
		return err
	}
	contextName := "AGENTS.md"
	if adapter == "claude_code" {
		contextName = "CLAUDE.md"
	}
	skills := a.SkillsDir
	if skills == "" {
		skills = ".agent-context/skills"
	}
	files := map[string]bool{contextName: true}
	if managed {
		files = map[string]bool{a.RootTarget: true}
		for _, seed := range a.Seeds {
			files[seed] = true
		}
		if adapter != "muse" {
			files[".agent-context/system-prompt.md"] = true
		}
		switch adapter {
		case "claude_code":
			files[".agent-context/mcp/claude_code.json"] = true
		case "codex_cli":
			files["curator-mcp.config.toml"] = true
		case "opencode":
			files["opencode.json"] = true
			files[".agent-context/mcp/opencode.json"] = true
		case "pi":
			files["SYSTEM.md"] = true
			files["APPEND_SYSTEM.md"] = true
		}
	}
	parents := map[string]bool{}
	for parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(skills))); parent != "."; parent = filepath.ToSlash(filepath.Dir(filepath.FromSlash(parent))) {
		parents[parent] = true
	}
	for file := range files {
		for parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(file))); parent != "."; parent = filepath.ToSlash(filepath.Dir(filepath.FromSlash(parent))) {
			parents[parent] = true
		}
	}
	if managed && (adapter == "claude_code" || adapter == "opencode") {
		parents[".agent-context"] = true
		parents[".agent-context/modules"] = true
	}
	var scan func(string, string) error
	scan = func(dir, rel string) error {
		before, err := ObserveDirectory(dir)
		if err != nil || before.Kind != "directory" {
			return invalid("backup directory identity/type: %v", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			next := e.Name()
			if rel != "" {
				next = rel + "/" + next
			}
			info, err := os.Lstat(filepath.Join(dir, e.Name()))
			if err != nil {
				return err
			}
			switch {
			case files[next]:
				if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
					return invalid("backup context type")
				}
			case next == skills || parents[next] || (skills == ".agent-context/skills" && next == ".agent-context"):
				if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return invalid("backup skill parent type")
				}
				if err := scan(filepath.Join(dir, e.Name()), next); err != nil {
					return err
				}
			case rel == skills && identifier.MatchString(e.Name()):
				if !info.IsDir() && !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
					return invalid("backup skill entry type")
				}
			case managed && (adapter == "claude_code" || adapter == "opencode") && strings.HasPrefix(next, ".agent-context/modules/"):
				parts := strings.Split(strings.TrimPrefix(next, ".agent-context/modules/"), "/")
				if !identifier.MatchString(parts[0]) {
					return invalid("backup module owner grammar")
				}
				if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
					if err := scan(filepath.Join(dir, e.Name()), next); err != nil {
						return err
					}
				} else if len(parts) < 2 || !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
					return invalid("backup module entry type")
				}
			default:
				return invalid("backup contains an unowned slot")
			}
		}
		after, err := ObserveDirectory(dir)
		if err != nil || after != before {
			return invalid("backup directory substituted: %v", err)
		}
		return nil
	}
	return scan(path, "")
}
