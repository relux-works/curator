// Package ui renders installed state as a terminal view (plan P11).
//
// The state machine is a pure reducer; the bubbletea model is a thin shell,
// so behavior tests run without a terminal (tuitestkit closed loop).
package ui

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/marker"
)

// SkillRow is one installed skill in the view.
type SkillRow struct {
	Name        string
	Ref         string
	Commit      string
	Context     bool
	Commands    []string
	Attestation string // "registry/status" or ""
	Substituted bool
}

// ProjectRow is one project with its installed skills.
type ProjectRow struct {
	Alias  string
	Path   string
	Skills []SkillRow
}

// Screen identifies the visible screen.
type Screen int

const (
	ScreenProjects Screen = iota
	ScreenSkills
)

// State is the whole UI state.
type State struct {
	Projects []ProjectRow
	Screen   Screen
	Cursor   int
	Selected int // project index when Screen == ScreenSkills
	Quitting bool
}

// Action is a reducer input.
type Action int

const (
	ActionUp Action = iota
	ActionDown
	ActionEnter
	ActionBack
	ActionQuit
)

// Reduce is the pure state transition.
func Reduce(state State, action Action) State {
	switch action {
	case ActionQuit:
		state.Quitting = true
	case ActionUp:
		if state.Cursor > 0 {
			state.Cursor--
		}
	case ActionDown:
		if state.Cursor < listLength(state)-1 {
			state.Cursor++
		}
	case ActionEnter:
		if state.Screen == ScreenProjects && len(state.Projects) > 0 {
			state.Selected = state.Cursor
			state.Screen = ScreenSkills
			state.Cursor = 0
		}
	case ActionBack:
		if state.Screen == ScreenSkills {
			state.Screen = ScreenProjects
			state.Cursor = state.Selected
		}
	}
	return state
}

func listLength(state State) int {
	if state.Screen == ScreenProjects {
		return len(state.Projects)
	}
	if state.Selected < len(state.Projects) {
		return len(state.Projects[state.Selected].Skills)
	}
	return 0
}

// LoadState reads projects and their install markers.
func LoadState(cfg *config.Config) State {
	var projects []ProjectRow
	var aliases []string
	for alias := range cfg.Projects {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		project := cfg.Projects[alias]
		projects = append(projects, ProjectRow{
			Alias:  alias,
			Path:   project.Path,
			Skills: skillsUnder(filepath.Join(project.Path, ".agents", "skills")),
		})
	}
	return State{Projects: projects}
}

func skillsUnder(skillsDir string) []SkillRow {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}
	var rows []SkillRow
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		recorded := marker.Read(filepath.Join(skillsDir, entry.Name()))
		if recorded == nil {
			continue
		}
		row := SkillRow{
			Name:        recorded.Name,
			Ref:         recorded.RefKind + " " + recorded.Ref,
			Commit:      shortCommit(recorded.Commit),
			Substituted: recorded.Substituted != "",
		}
		if recorded.Activation != nil {
			row.Context = recorded.Activation.Context
			row.Commands = recorded.Activation.Commands
		}
		if recorded.Attestation != nil {
			row.Attestation = recorded.Attestation.Registry + "/" + recorded.Attestation.Status
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}

func shortCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}
