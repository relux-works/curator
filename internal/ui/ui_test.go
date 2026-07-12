package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/relux-works/skill-go-testing-tools/tuitestkit"
)

func fixtureState() State {
	return State{
		Projects: []ProjectRow{
			{Alias: "alpha", Path: "/work/alpha", Skills: []SkillRow{
				{Name: "skill-a", Ref: "tag v1", Commit: "abc1234", Context: true,
					Commands: []string{"a-tool"}, Attestation: "corp/audited"},
				{Name: "skill-b", Ref: "tag v2", Commit: "def5678", Context: false, Substituted: true},
			}},
			{Alias: "beta", Path: "/work/beta"},
		},
	}
}

func TestReducerNavigation(t *testing.T) {
	tuitestkit.RunReducerTests(t, Reduce, []tuitestkit.ReducerTest[State, Action]{
		{
			Name: "down moves the cursor", Initial: fixtureState(), Action: ActionDown,
			Assert: func(t *testing.T, got State) {
				if got.Cursor != 1 {
					t.Fatalf("cursor = %d", got.Cursor)
				}
			},
		},
		{
			Name:    "down clamps at the end",
			Initial: func() State { s := fixtureState(); s.Cursor = 1; return s }(),
			Action:  ActionDown,
			Assert: func(t *testing.T, got State) {
				if got.Cursor != 1 {
					t.Fatalf("cursor overflowed: %d", got.Cursor)
				}
			},
		},
		{
			Name: "up clamps at zero", Initial: fixtureState(), Action: ActionUp,
			Assert: func(t *testing.T, got State) {
				if got.Cursor != 0 {
					t.Fatalf("cursor underflowed: %d", got.Cursor)
				}
			},
		},
		{
			Name: "enter opens the project", Initial: fixtureState(), Action: ActionEnter,
			Assert: func(t *testing.T, got State) {
				if got.Screen != ScreenSkills || got.Selected != 0 || got.Cursor != 0 {
					t.Fatalf("state: %+v", got)
				}
			},
		},
		{
			Name: "quit flags quitting", Initial: fixtureState(), Action: ActionQuit,
			Assert: func(t *testing.T, got State) {
				if !got.Quitting {
					t.Fatal("not quitting")
				}
			},
		},
	})
}

func TestReducerRoundTrip(t *testing.T) {
	tuitestkit.RunReducerSequences(t, Reduce, []tuitestkit.ReducerSequence[State, Action]{
		{
			Name:    "open skills, move, back restores the project cursor",
			Initial: fixtureState(),
			Steps: []tuitestkit.Step[State, Action]{
				{Name: "down to beta", Action: ActionDown},
				{Name: "up back to alpha", Action: ActionUp},
				{Name: "enter alpha", Action: ActionEnter},
				{Name: "down inside skills", Action: ActionDown, Assert: func(t *testing.T, got State) {
					if got.Screen != ScreenSkills || got.Cursor != 1 {
						t.Fatalf("state: %+v", got)
					}
				}},
				{Name: "back to projects", Action: ActionBack, Assert: func(t *testing.T, got State) {
					if got.Screen != ScreenProjects || got.Cursor != 0 {
						t.Fatalf("state: %+v", got)
					}
				}},
			},
		},
	})
}

func TestGoldenScreens(t *testing.T) {
	projects := NewModel(fixtureState())
	tuitestkit.SnapshotView(t, projects, "projects")

	opened := NewModel(Reduce(fixtureState(), ActionEnter))
	tuitestkit.SnapshotView(t, opened, "skills")
}

func TestKeysDriveTheModel(t *testing.T) {
	model := NewModel(fixtureState())
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if next.(Model).State.Cursor != 1 {
		t.Fatal("j must move down")
	}
	next, cmd := next.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !next.(Model).State.Quitting || cmd == nil {
		t.Fatal("q must quit")
	}
	if view := next.(Model).View(); view != "" {
		t.Fatalf("quitting view must be empty: %q", view)
	}
}
