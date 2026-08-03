package picker_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vicgarci/sadb/internal/picker"
)

// pressKey drives the model's Update method with a synthetic key press.
func pressKey(m picker.Model, key string) picker.Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	switch key {
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	}
	updated, _ := m.Update(msg)
	return updated.(picker.Model)
}

func TestModel_InitialState(t *testing.T) {
	m := picker.New([]string{"emulator-5554", "192.168.1.1:5555"})

	if m.Cursor() != 0 {
		t.Errorf("expected initial cursor 0, got %d", m.Cursor())
	}
	if m.Selected() != "" {
		t.Errorf("expected no selection initially, got %q", m.Selected())
	}
	if m.Aborted() {
		t.Error("expected Aborted=false initially")
	}
}

func TestModel_NavigateDown(t *testing.T) {
	m := picker.New([]string{"a", "b", "c"})

	m = pressKey(m, "down")
	if m.Cursor() != 1 {
		t.Errorf("expected cursor 1 after down, got %d", m.Cursor())
	}

	m = pressKey(m, "down")
	if m.Cursor() != 2 {
		t.Errorf("expected cursor 2 after second down, got %d", m.Cursor())
	}
}

func TestModel_NavigateDown_Wraps(t *testing.T) {
	m := picker.New([]string{"a", "b"})

	m = pressKey(m, "down")
	m = pressKey(m, "down") // should wrap back to 0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to wrap to 0, got %d", m.Cursor())
	}
}

func TestModel_NavigateUp(t *testing.T) {
	m := picker.New([]string{"a", "b", "c"})
	m = pressKey(m, "down") // cursor=1
	m = pressKey(m, "up")   // cursor=0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", m.Cursor())
	}
}

func TestModel_NavigateUp_Wraps(t *testing.T) {
	m := picker.New([]string{"a", "b"})
	m = pressKey(m, "up") // should wrap to last item
	if m.Cursor() != 1 {
		t.Errorf("expected cursor to wrap to 1, got %d", m.Cursor())
	}
}

func TestModel_SelectWithEnter(t *testing.T) {
	m := picker.New([]string{"emulator-5554", "192.168.1.1:5555"})
	m = pressKey(m, "down")  // move to second item
	m = pressKey(m, "enter") // select it

	if m.Selected() != "192.168.1.1:5555" {
		t.Errorf("expected selected %q, got %q", "192.168.1.1:5555", m.Selected())
	}
	if m.Aborted() {
		t.Error("expected Aborted=false after enter")
	}
}

func TestModel_AbortWithCtrlC(t *testing.T) {
	m := picker.New([]string{"a", "b"})
	m = pressKey(m, "ctrl+c")

	if !m.Aborted() {
		t.Error("expected Aborted=true after ctrl+c")
	}
	if m.Selected() != "" {
		t.Errorf("expected no selection after ctrl+c, got %q", m.Selected())
	}
}

func TestModel_AbortWithEsc(t *testing.T) {
	m := picker.New([]string{"a", "b"})
	m = pressKey(m, "esc")

	if !m.Aborted() {
		t.Error("expected Aborted=true after esc")
	}
}

func TestModel_ViewContainsDevices(t *testing.T) {
	devices := []string{"emulator-5554", "192.168.1.1:5555"}
	m := picker.New(devices)

	view := m.View()
	for _, d := range devices {
		if !containsSubstring(view, d) {
			t.Errorf("expected view to contain device %q\nview:\n%s", d, view)
		}
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

