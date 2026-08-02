// Package picker provides an interactive Bubble Tea TUI for selecting an Android device.
package picker

import (
	"errors"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrAborted is returned when the user cancels the picker with Ctrl+C or Escape.
var ErrAborted = errors.New("picker aborted")

// Model is the Bubble Tea model for device selection.
// It is exported so tests can drive it without running a full terminal program.
type Model struct {
	devices  []string
	cursor   int
	aborted  bool
	selected string
}

// New creates a Model with the given device list. Cursor starts at 0.
func New(devices []string) Model {
	return Model{devices: devices}
}

// Cursor returns the current highlighted index.
func (m Model) Cursor() int { return m.cursor }

// Selected returns the serial of the chosen device, or "" if no selection has been made.
func (m Model) Selected() string { return m.selected }

// Aborted reports whether the user cancelled.
func (m Model) Aborted() bool { return m.aborted }

// Init satisfies tea.Model; no startup commands needed.
func (m Model) Init() tea.Cmd { return nil }

// Update processes keyboard input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.cursor = (m.cursor - 1 + len(m.devices)) % len(m.devices)
		case tea.KeyDown:
			m.cursor = (m.cursor + 1) % len(m.devices)
		case tea.KeyEnter:
			m.selected = m.devices[m.cursor]
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the device list with a cursor indicator.
func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString("Select a device:\n\n")
	for i, d := range m.devices {
		if i == m.cursor {
			sb.WriteString(fmt.Sprintf("  > %s\n", d))
		} else {
			sb.WriteString(fmt.Sprintf("    %s\n", d))
		}
	}
	sb.WriteString("\n  ↑/↓ navigate  enter select  ctrl+c/esc cancel\n")
	return sb.String()
}

// BubbleTeaPicker is a real Picker implementation that runs the interactive TUI.
type BubbleTeaPicker struct {
	// Stderr is the writer for TUI output. Should be set to os.Stderr by callers.
	Stderr io.Writer
}

// Pick launches the Bubble Tea program and returns the selected serial.
// Returns ErrAborted if the user presses Ctrl+C or Escape.
func (p BubbleTeaPicker) Pick(devices []string) (string, error) {
	out := p.Stderr
	if out == nil {
		out = io.Discard
	}
	prog := tea.NewProgram(New(devices), tea.WithOutput(out))

	result, err := prog.Run()
	if err != nil {
		return "", fmt.Errorf("device picker: %w", err)
	}

	m := result.(Model)
	if m.Aborted() {
		return "", ErrAborted
	}
	return m.Selected(), nil
}
