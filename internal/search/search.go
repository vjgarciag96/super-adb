// Package search provides an interactive Bubble Tea TUI for selecting an installed package.
package search

import (
	"errors"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/viewport"
)

// ErrAborted is returned when the user cancels the search with Ctrl+C or Escape.
var ErrAborted = errors.New("search aborted")

// ParsePackages parses the output of `adb shell pm list packages` into a sorted slice of package names.
// Each line in the output is expected to be in the format "package:<packagename>".
func ParsePackages(output string) []string {
	var packages []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "package:"); ok && after != "" {
			packages = append(packages, after)
		}
	}
	return packages
}

// FetchPackages runs `adb shell pm list packages` on the given device and returns the package list.
func FetchPackages(serial string, runner adb.Runner) ([]string, error) {
	out, err := runner.Run(serial, "shell", "pm", "list", "packages")
	if err != nil {
		return nil, fmt.Errorf("fetch packages: %w", err)
	}
	return ParsePackages(out), nil
}

// Model is the Bubble Tea model for package search.
// It is exported so tests can drive it without running a full terminal program.
type Model struct {
	packages []string
	filter   string
	vp       viewport.Viewport
	aborted  bool
	selected string
}

// New creates a Model with the given package list.
func New(packages []string) Model {
	return Model{packages: packages, vp: viewport.Viewport{Height: 10}}
}

// Filter returns the current filter string.
func (m Model) Filter() string { return m.filter }

// Cursor returns the current highlighted index within the filtered list.
func (m Model) Cursor() int { return m.vp.Cursor }

// ViewOffset returns the index of the first visible item in the filtered list.
func (m Model) ViewOffset() int { return m.vp.Offset }

// Selected returns the chosen package name, or "" if no selection has been made.
func (m Model) Selected() string { return m.selected }

// Aborted reports whether the user cancelled.
func (m Model) Aborted() bool { return m.aborted }

// filtered returns the subset of packages matching the current filter.
func (m Model) filtered() []string {
	if m.filter == "" {
		return m.packages
	}
	var result []string
	lower := strings.ToLower(m.filter)
	for _, p := range m.packages {
		if strings.Contains(strings.ToLower(p), lower) {
			result = append(result, p)
		}
	}
	return result
}

// Init satisfies tea.Model; no startup commands needed.
func (m Model) Init() tea.Cmd { return nil }

// Update processes keyboard input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp.SetHeight(msg.Height, 6, 5)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.vp.MoveUp(len(m.filtered()))
		case tea.KeyDown:
			m.vp.MoveDown(len(m.filtered()))
		case tea.KeyEnter:
			filtered := m.filtered()
			if len(filtered) > 0 {
				m.selected = filtered[m.vp.Cursor]
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit
		case tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.vp.Cursor = 0
				m.vp.Offset = 0
			}
		case tea.KeyRunes:
			m.filter += string(msg.Runes)
			m.vp.Cursor = 0
			m.vp.Offset = 0
		}
	}
	return m, nil
}

// View renders the package list with the current filter and a cursor indicator.
// Only a window of height items is rendered at a time; scroll hints appear when
// more items exist above or below the visible range.
func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search packages: %s\n\n", m.filter))

	filtered := m.filtered()
	if len(filtered) == 0 {
		sb.WriteString("  (no matches)\n")
	} else {
		m.vp.RenderList(&sb, filtered)
	}
	sb.WriteString("\n  type to filter  ↑/↓ navigate  enter select  ctrl+c/esc cancel\n")
	return sb.String()
}

// Selector interactively selects a package from the given list.
// It returns ErrAborted if the user cancels.
type Selector interface {
	Select(packages []string) (string, error)
}

// BubbleTeaSelector runs the interactive package search TUI against a pre-fetched package list.
type BubbleTeaSelector struct {
	// Stderr is the writer for TUI output. Should be set to os.Stderr by callers.
	Stderr io.Writer
}

// Select opens the interactive TUI with the given package list and returns the
// selected package name, or ErrAborted if the user cancels.
func (s BubbleTeaSelector) Select(packages []string) (string, error) {
	out := s.Stderr
	if out == nil {
		out = io.Discard
	}
	prog := tea.NewProgram(New(packages), tea.WithOutput(out))

	result, err := prog.Run()
	if err != nil {
		return "", fmt.Errorf("package search: %w", err)
	}

	m := result.(Model)
	if m.Aborted() {
		return "", ErrAborted
	}
	return m.Selected(), nil
}

// BubbleTeaSearcher fetches the package list from the active device and runs the interactive TUI.
type BubbleTeaSearcher struct {
	// Stderr is the writer for TUI output. Should be set to os.Stderr by callers.
	Stderr io.Writer
}

// Search fetches installed packages from the device identified by serial and launches the
// interactive TUI. It returns the selected package name, or ErrAborted if the user cancels.
func (s BubbleTeaSearcher) Search(serial string, runner adb.Runner) (string, error) {
	packages, err := FetchPackages(serial, runner)
	if err != nil {
		return "", err
	}
	return BubbleTeaSelector{Stderr: s.Stderr}.Select(packages)
}
