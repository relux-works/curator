package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	attestedStyle = lipgloss.NewStyle().Bold(true)
)

// Model is the bubbletea shell around the pure reducer.
type Model struct {
	State State
}

// NewModel builds the model from loaded state.
func NewModel(state State) Model { return Model{State: state} }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model: keys map to actions, the reducer does the rest.
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q", "ctrl+c":
		m.State = Reduce(m.State, ActionQuit)
		return m, tea.Quit
	case "up", "k":
		m.State = Reduce(m.State, ActionUp)
	case "down", "j":
		m.State = Reduce(m.State, ActionDown)
	case "enter", "l":
		m.State = Reduce(m.State, ActionEnter)
	case "esc", "h":
		m.State = Reduce(m.State, ActionBack)
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.State.Quitting {
		return ""
	}
	if m.State.Screen == ScreenSkills {
		return m.skillsView()
	}
	return m.projectsView()
}

func (m Model) projectsView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("curator: projects") + "\n\n")
	if len(m.State.Projects) == 0 {
		b.WriteString(dimStyle.Render("no registered projects") + "\n")
	}
	for index, project := range m.State.Projects {
		prefix := "  "
		line := fmt.Sprintf("%s  %s (%d skills)", project.Alias, dimStyle.Render(project.Path), len(project.Skills))
		if index == m.State.Cursor {
			prefix = cursorStyle.Render("> ")
			line = cursorStyle.Render(fmt.Sprintf("%s  %s (%d skills)", project.Alias, project.Path, len(project.Skills)))
		}
		b.WriteString(prefix + line + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("j/k move · enter open · q quit") + "\n")
	return b.String()
}

func (m Model) skillsView() string {
	project := m.State.Projects[m.State.Selected]
	var b strings.Builder
	b.WriteString(titleStyle.Render("curator: "+project.Alias) + "\n\n")
	if len(project.Skills) == 0 {
		b.WriteString(dimStyle.Render("nothing installed") + "\n")
	}
	for index, skill := range project.Skills {
		marker := "  "
		if index == m.State.Cursor {
			marker = cursorStyle.Render("> ")
		}
		context := "context:no"
		if skill.Context {
			context = "context:yes"
		}
		suffix := ""
		if skill.Attestation != "" {
			suffix += " " + attestedStyle.Render("["+skill.Attestation+"]")
		}
		if skill.Substituted {
			suffix += " " + dimStyle.Render("SUBSTITUTED")
		}
		commands := ""
		if len(skill.Commands) > 0 {
			commands = " commands:" + strings.Join(skill.Commands, ",")
		}
		fmt.Fprintf(&b, "%s%s  %s %s  %s%s%s\n",
			marker, skill.Name, skill.Ref, skill.Commit, context, commands, suffix)
	}
	b.WriteString("\n" + dimStyle.Render("j/k move · esc back · q quit") + "\n")
	return b.String()
}
