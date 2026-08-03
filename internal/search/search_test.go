package search_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/search"
)

// pressKey drives the model's Update method with a synthetic key press.
func pressKey(m search.Model, key string) search.Model {
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
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	}
	updated, _ := m.Update(msg)
	return updated.(search.Model)
}

// typeString simulates typing a string one rune at a time.
func typeString(m search.Model, s string) search.Model {
	for _, r := range s {
		m = pressKey(m, string(r))
	}
	return m
}

// --- ParsePackages tests ---

func TestParsePackages_StandardOutput(t *testing.T) {
	output := "package:com.example.foo\npackage:com.example.bar\npackage:org.test.baz\n"
	got := search.ParsePackages(output)
	want := []string{"com.example.foo", "com.example.bar", "org.test.baz"}
	if len(got) != len(want) {
		t.Fatalf("expected %d packages, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestParsePackages_EmptyOutput(t *testing.T) {
	got := search.ParsePackages("")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestParsePackages_SkipsNonPackageLines(t *testing.T) {
	output := "WARNING: linker: libdvm.so has text relocations.\npackage:com.example.app\n"
	got := search.ParsePackages(output)
	if len(got) != 1 || got[0] != "com.example.app" {
		t.Errorf("expected [com.example.app], got %v", got)
	}
}

// --- FetchPackages tests ---

func TestFetchPackages_MakesCorrectADBCall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil)

	packages, err := search.FetchPackages("emulator-5554", f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call, got %d", len(f.Calls))
	}
	call := f.Calls[0]
	if call.Serial != "emulator-5554" {
		t.Errorf("expected serial %q, got %q", "emulator-5554", call.Serial)
	}
	wantArgs := []string{"shell", "pm", "list", "packages"}
	if len(call.Args) != len(wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, call.Args)
	}
	for i, a := range wantArgs {
		if call.Args[i] != a {
			t.Errorf("arg[%d]: expected %q, got %q", i, a, call.Args[i])
		}
	}

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d: %v", len(packages), packages)
	}
}

// --- Model tests ---

func TestModel_InitialState(t *testing.T) {
	m := search.New([]string{"com.example.foo", "com.example.bar"})

	if m.Cursor() != 0 {
		t.Errorf("expected initial cursor 0, got %d", m.Cursor())
	}
	if m.Selected() != "" {
		t.Errorf("expected no selection initially, got %q", m.Selected())
	}
	if m.Aborted() {
		t.Error("expected Aborted=false initially")
	}
	if m.Filter() != "" {
		t.Errorf("expected empty filter initially, got %q", m.Filter())
	}
}

func TestModel_NavigateDown(t *testing.T) {
	m := search.New([]string{"a", "b", "c"})

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
	m := search.New([]string{"a", "b"})
	m = pressKey(m, "down")
	m = pressKey(m, "down") // should wrap back to 0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to wrap to 0, got %d", m.Cursor())
	}
}

func TestModel_NavigateUp_Wraps(t *testing.T) {
	m := search.New([]string{"a", "b"})
	m = pressKey(m, "up") // should wrap to last item
	if m.Cursor() != 1 {
		t.Errorf("expected cursor to wrap to 1, got %d", m.Cursor())
	}
}

func TestModel_SelectWithEnter(t *testing.T) {
	m := search.New([]string{"com.example.foo", "com.example.bar"})
	m = pressKey(m, "down")  // move to second item
	m = pressKey(m, "enter") // select it

	if m.Selected() != "com.example.bar" {
		t.Errorf("expected selected %q, got %q", "com.example.bar", m.Selected())
	}
	if m.Aborted() {
		t.Error("expected Aborted=false after enter")
	}
}

func TestModel_AbortWithCtrlC(t *testing.T) {
	m := search.New([]string{"a", "b"})
	m = pressKey(m, "ctrl+c")

	if !m.Aborted() {
		t.Error("expected Aborted=true after ctrl+c")
	}
	if m.Selected() != "" {
		t.Errorf("expected no selection after ctrl+c, got %q", m.Selected())
	}
}

func TestModel_AbortWithEsc(t *testing.T) {
	m := search.New([]string{"a", "b"})
	m = pressKey(m, "esc")

	if !m.Aborted() {
		t.Error("expected Aborted=true after esc")
	}
}

func TestModel_TypingFiltersPackages(t *testing.T) {
	m := search.New([]string{"com.example.foo", "com.example.bar", "org.other.baz"})
	m = typeString(m, "bar")

	if m.Filter() != "bar" {
		t.Errorf("expected filter %q, got %q", "bar", m.Filter())
	}
	// Only "com.example.bar" should match; cursor should be 0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0 in filtered list, got %d", m.Cursor())
	}
}

func TestModel_FilterResetsCursor(t *testing.T) {
	m := search.New([]string{"com.alpha", "com.beta", "com.gamma"})
	m = pressKey(m, "down") // cursor = 1
	m = typeString(m, "a")  // filter narrows list; cursor should reset to 0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0 after typing, got %d", m.Cursor())
	}
}

func TestModel_SelectFromFilteredList(t *testing.T) {
	m := search.New([]string{"com.example.foo", "com.example.bar", "org.other.baz"})
	m = typeString(m, "bar")
	m = pressKey(m, "enter")

	if m.Selected() != "com.example.bar" {
		t.Errorf("expected selected %q, got %q", "com.example.bar", m.Selected())
	}
}

func TestModel_BackspaceRemovesFilterChar(t *testing.T) {
	m := search.New([]string{"com.example.foo"})
	m = typeString(m, "fo")
	m = pressKey(m, "backspace")

	if m.Filter() != "f" {
		t.Errorf("expected filter %q after backspace, got %q", "f", m.Filter())
	}
}

func TestModel_ViewContainsPackages(t *testing.T) {
	packages := []string{"com.example.foo", "com.example.bar"}
	m := search.New(packages)

	view := m.View()
	for _, p := range packages {
		if !strings.Contains(view, p) {
			t.Errorf("expected view to contain package %q\nview:\n%s", p, view)
		}
	}
}

func TestModel_ViewShowsNoMatchesWhenFilterExcludesAll(t *testing.T) {
	m := search.New([]string{"com.example.foo", "com.example.bar"})
	m = typeString(m, "zzz")

	view := m.View()
	if !strings.Contains(view, "no matches") {
		t.Errorf("expected view to show 'no matches'\nview:\n%s", view)
	}
}

func TestModel_ViewShowsFilter(t *testing.T) {
	m := search.New([]string{"com.example.foo"})
	m = typeString(m, "foo")

	view := m.View()
	if !strings.Contains(view, "foo") {
		t.Errorf("expected view to contain filter text %q\nview:\n%s", "foo", view)
	}
}

// --- Viewport / scrolling tests ---

// makePackages returns n distinct package names for use in viewport tests.
func makePackages(n int) []string {
	pkgs := make([]string, n)
	for i := range pkgs {
		pkgs[i] = fmt.Sprintf("com.example.pkg%03d", i)
	}
	return pkgs
}

func TestModel_ViewOffset_InitiallyZero(t *testing.T) {
	m := search.New(makePackages(20))
	if m.ViewOffset() != 0 {
		t.Errorf("expected initial viewOffset 0, got %d", m.ViewOffset())
	}
}

func TestModel_ViewOffset_ShiftsDownWhenCursorLeavesWindow(t *testing.T) {
	// Default height is 10; pressing down 10 times moves cursor to index 10,
	// which is one past the last visible item [0..9], so viewOffset must advance.
	m := search.New(makePackages(20))
	for i := 0; i < 10; i++ {
		m = pressKey(m, "down")
	}
	if m.Cursor() != 10 {
		t.Fatalf("expected cursor 10, got %d", m.Cursor())
	}
	if m.ViewOffset() == 0 {
		t.Errorf("expected viewOffset to have advanced past 0, got %d", m.ViewOffset())
	}
	// cursor must be within the visible window
	if m.Cursor() < m.ViewOffset() || m.Cursor() >= m.ViewOffset()+10 {
		t.Errorf("cursor %d outside visible window [%d, %d)", m.Cursor(), m.ViewOffset(), m.ViewOffset()+10)
	}
}

func TestModel_ViewOffset_ShiftsUpWhenCursorLeavesWindow(t *testing.T) {
	// Navigate to the bottom of a 20-item list, then navigate back up past the window top.
	m := search.New(makePackages(20))
	for i := 0; i < 15; i++ {
		m = pressKey(m, "down")
	}
	offsetAtBottom := m.ViewOffset()

	for i := 0; i < 10; i++ {
		m = pressKey(m, "up")
	}
	if m.ViewOffset() >= offsetAtBottom {
		t.Errorf("expected viewOffset to decrease after navigating up; was %d, now %d", offsetAtBottom, m.ViewOffset())
	}
	if m.Cursor() < m.ViewOffset() || m.Cursor() >= m.ViewOffset()+10 {
		t.Errorf("cursor %d outside visible window [%d, %d)", m.Cursor(), m.ViewOffset(), m.ViewOffset()+10)
	}
}

func TestModel_ViewOffset_ResetsOnWrapToTop(t *testing.T) {
	// Navigate to end of list so viewOffset > 0, then wrap to top with one more down.
	m := search.New(makePackages(15))
	for i := 0; i < 14; i++ {
		m = pressKey(m, "down")
	}
	if m.ViewOffset() == 0 {
		t.Fatal("expected viewOffset > 0 near end of list")
	}
	m = pressKey(m, "down") // wraps to index 0
	if m.Cursor() != 0 {
		t.Fatalf("expected cursor to wrap to 0, got %d", m.Cursor())
	}
	if m.ViewOffset() != 0 {
		t.Errorf("expected viewOffset to reset to 0 on wrap, got %d", m.ViewOffset())
	}
}

func TestModel_ViewOffset_ShowsLastItemsOnWrapToBottom(t *testing.T) {
	// From the top, pressing up wraps cursor to the last item.
	// viewOffset must shift so the last item is visible.
	m := search.New(makePackages(15))
	m = pressKey(m, "up") // wraps to index 14
	if m.Cursor() != 14 {
		t.Fatalf("expected cursor 14 after wrap, got %d", m.Cursor())
	}
	if m.Cursor() < m.ViewOffset() || m.Cursor() >= m.ViewOffset()+10 {
		t.Errorf("cursor %d outside visible window [%d, %d)", m.Cursor(), m.ViewOffset(), m.ViewOffset()+10)
	}
}

func TestModel_ViewOffset_ResetsOnTyping(t *testing.T) {
	m := search.New(makePackages(20))
	for i := 0; i < 15; i++ {
		m = pressKey(m, "down")
	}
	if m.ViewOffset() == 0 {
		t.Fatal("expected viewOffset > 0 before typing")
	}
	m = typeString(m, "x") // no matches, but offset must reset
	if m.ViewOffset() != 0 {
		t.Errorf("expected viewOffset 0 after typing, got %d", m.ViewOffset())
	}
}

func TestModel_ViewOffset_ResetsOnBackspace(t *testing.T) {
	m := search.New(makePackages(20))
	m = typeString(m, "com") // filter narrows; navigate down
	for i := 0; i < 15; i++ {
		m = pressKey(m, "down")
	}
	if m.ViewOffset() == 0 {
		t.Fatal("expected viewOffset > 0 before backspace")
	}
	m = pressKey(m, "backspace")
	if m.ViewOffset() != 0 {
		t.Errorf("expected viewOffset 0 after backspace, got %d", m.ViewOffset())
	}
}

func TestModel_WindowSizeMsg_UpdatesHeight(t *testing.T) {
	m := search.New(makePackages(5))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = updated.(search.Model)
	// height = 30 - 6 = 24
	// We verify indirectly: with 5 packages and height 24, the view must show all 5.
	view := m.View()
	for i := 0; i < 5; i++ {
		pkg := fmt.Sprintf("com.example.pkg%03d", i)
		if !strings.Contains(view, pkg) {
			t.Errorf("expected view to contain %q after WindowSizeMsg\nview:\n%s", pkg, view)
		}
	}
}

func TestModel_WindowSizeMsg_MinimumHeight(t *testing.T) {
	m := search.New(makePackages(5))
	// Very small terminal: height - 6 would be negative, must clamp to 5.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 3})
	m = updated.(search.Model)
	view := m.View()
	// Should not panic and should still render something.
	if !strings.Contains(view, "Search packages") {
		t.Errorf("expected view header after tiny WindowSizeMsg\nview:\n%s", view)
	}
}

func TestModel_View_ShowsScrollHintBelow(t *testing.T) {
	// 20 items, default height 10: items 10-19 are below the window.
	m := search.New(makePackages(20))
	view := m.View()
	if !strings.Contains(view, "more below") {
		t.Errorf("expected 'more below' hint in view\nview:\n%s", view)
	}
}

func TestModel_View_ShowsScrollHintAbove(t *testing.T) {
	m := search.New(makePackages(20))
	for i := 0; i < 10; i++ {
		m = pressKey(m, "down")
	}
	view := m.View()
	if !strings.Contains(view, "more above") {
		t.Errorf("expected 'more above' hint after scrolling down\nview:\n%s", view)
	}
}

func TestModel_View_NoScrollHintsWhenAllFit(t *testing.T) {
	// 5 items fit within the default height of 10: no hints expected.
	m := search.New(makePackages(5))
	view := m.View()
	if strings.Contains(view, "more above") || strings.Contains(view, "more below") {
		t.Errorf("expected no scroll hints when all items fit\nview:\n%s", view)
	}
}

func TestModel_View_OnlyRendersVisibleItems(t *testing.T) {
	// With height=10 and 20 items, only items 0-9 should appear initially.
	m := search.New(makePackages(20))
	view := m.View()
	if !strings.Contains(view, "pkg000") {
		t.Errorf("expected first item in view")
	}
	if strings.Contains(view, "pkg010") {
		t.Errorf("did not expect item 10 in initial view (height=10)")
	}
}
