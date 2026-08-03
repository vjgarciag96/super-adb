// Package picker provides an interactive Bubble Tea TUI for selecting an Android device.
package picker

import (
	"errors"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vicgarci/sadb/internal/viewport"
)

// ErrAborted is returned when the user cancels the picker with Ctrl+C or Escape.
var ErrAborted = errors.New("picker aborted")

// Model is the Bubble Tea model for device selection.
// It is exported so tests can drive it without running a full terminal program.
type Model struct {
	devices  []string
	vp       viewport.Viewport
	aborted  bool
	selected string
}

// New creates a Model with the given device list. Cursor starts at 0.
func New(devices []string) Model {
	return Model{devices: devices, vp: viewport.Viewport{Height: 10}}
}

// Cursor returns the current highlighted index.
func (m Model) Cursor() int { return m.vp.Cursor }

// ViewOffset returns the index of the first visible device in the list.
func (m Model) ViewOffset() int { return m.vp.Offset }

// Selected returns the serial of the chosen device, or "" if no selection has been made.
func (m Model) Selected() string { return m.selected }

// Aborted reports whether the user cancelled.
func (m Model) Aborted() bool { return m.aborted }

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
			m.vp.MoveUp(len(m.devices))
		case tea.KeyDown:
			m.vp.MoveDown(len(m.devices))
		case tea.KeyEnter:
			if len(m.devices) > 0 {
				m.selected = m.devices[m.vp.Cursor]
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the device list with a cursor indicator.
// Only a window of height items is rendered at a time; scroll hints appear
// when more items exist above or below the visible range.
func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString("Select a device:\n\n")

	if len(m.devices) == 0 {
		sb.WriteString("  (no devices)\n")
	} else {
		m.vp.RenderList(&sb, m.devices)
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
