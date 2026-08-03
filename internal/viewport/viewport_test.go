package viewport_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vicgarci/sadb/internal/viewport"
)

// --- SetHeight tests ---

func TestSetHeight_Normal(t *testing.T) {
	var vp viewport.Viewport
	vp.SetHeight(30, 6, 5)
	if vp.Height != 24 {
		t.Errorf("expected Height 24, got %d", vp.Height)
	}
}

func TestSetHeight_ClampsToMin(t *testing.T) {
	var vp viewport.Viewport
	vp.SetHeight(3, 6, 5) // 3-6 = -3, clamp to 5
	if vp.Height != 5 {
		t.Errorf("expected Height 5, got %d", vp.Height)
	}
}

func TestSetHeight_ExactMinBoundary(t *testing.T) {
	var vp viewport.Viewport
	vp.SetHeight(11, 6, 5) // 11-6 = 5, exactly at min
	if vp.Height != 5 {
		t.Errorf("expected Height 5, got %d", vp.Height)
	}
}

func TestSetHeight_OneLessThanMin(t *testing.T) {
	var vp viewport.Viewport
	vp.SetHeight(10, 6, 5) // 10-6 = 4, just below min
	if vp.Height != 5 {
		t.Errorf("expected Height 5, got %d", vp.Height)
	}
}

// --- Snap tests ---

func TestSnap_CursorAboveWindow(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 5, Height: 3}
	vp.Snap(20)
	if vp.Offset != 0 {
		t.Errorf("expected Offset 0, got %d", vp.Offset)
	}
}

func TestSnap_CursorBelowWindow(t *testing.T) {
	vp := viewport.Viewport{Cursor: 12, Offset: 0, Height: 10}
	vp.Snap(20)
	if vp.Offset != 3 { // 12 - 10 + 1
		t.Errorf("expected Offset 3, got %d", vp.Offset)
	}
}

func TestSnap_CursorInWindow_NoChange(t *testing.T) {
	vp := viewport.Viewport{Cursor: 5, Offset: 2, Height: 10}
	vp.Snap(20)
	if vp.Offset != 2 {
		t.Errorf("expected Offset unchanged at 2, got %d", vp.Offset)
	}
}

func TestSnap_OffsetNegativeClamp(t *testing.T) {
	// Cursor=0, Offset=1 -> cursor above window -> Offset set to Cursor=0, no negative.
	vp := viewport.Viewport{Cursor: 0, Offset: 1, Height: 5}
	vp.Snap(10)
	if vp.Offset != 0 {
		t.Errorf("expected Offset 0, got %d", vp.Offset)
	}
}

// --- VisibleRange tests ---

func TestVisibleRange_Normal(t *testing.T) {
	vp := viewport.Viewport{Offset: 2, Height: 5}
	start, end := vp.VisibleRange(20)
	if start != 2 || end != 7 {
		t.Errorf("expected (2, 7), got (%d, %d)", start, end)
	}
}

func TestVisibleRange_ClampsToTotal(t *testing.T) {
	vp := viewport.Viewport{Offset: 18, Height: 10}
	start, end := vp.VisibleRange(20)
	if start != 18 || end != 20 {
		t.Errorf("expected (18, 20), got (%d, %d)", start, end)
	}
}

func TestVisibleRange_ZeroHeight(t *testing.T) {
	vp := viewport.Viewport{Offset: 0, Height: 0}
	start, end := vp.VisibleRange(15)
	if start != 0 || end != 15 {
		t.Errorf("expected (0, 15), got (%d, %d)", start, end)
	}
}

func TestVisibleRange_NegativeHeight(t *testing.T) {
	vp := viewport.Viewport{Offset: 0, Height: -1}
	start, end := vp.VisibleRange(15)
	if start != 0 || end != 15 {
		t.Errorf("expected (0, 15), got (%d, %d)", start, end)
	}
}

// --- HiddenAbove tests ---

func TestHiddenAbove_Zero(t *testing.T) {
	vp := viewport.Viewport{Offset: 0}
	if vp.HiddenAbove() != 0 {
		t.Errorf("expected 0, got %d", vp.HiddenAbove())
	}
}

func TestHiddenAbove_Positive(t *testing.T) {
	vp := viewport.Viewport{Offset: 5}
	if vp.HiddenAbove() != 5 {
		t.Errorf("expected 5, got %d", vp.HiddenAbove())
	}
}

// --- HiddenBelow tests ---

func TestHiddenBelow_None(t *testing.T) {
	vp := viewport.Viewport{Offset: 0, Height: 20}
	if got := vp.HiddenBelow(10); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestHiddenBelow_Some(t *testing.T) {
	vp := viewport.Viewport{Offset: 0, Height: 10}
	if got := vp.HiddenBelow(20); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
}

func TestHiddenBelow_AtEnd(t *testing.T) {
	vp := viewport.Viewport{Offset: 15, Height: 10}
	// VisibleRange(20) -> end clamped to 20; 20-20 = 0
	if got := vp.HiddenBelow(20); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// --- MoveDown tests ---

func TestMoveDown_AdvancesCursor(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	vp.MoveDown(5)
	if vp.Cursor != 1 {
		t.Errorf("expected Cursor 1, got %d", vp.Cursor)
	}
}

func TestMoveDown_WrapsToTop(t *testing.T) {
	vp := viewport.Viewport{Cursor: 4, Offset: 0, Height: 10}
	vp.MoveDown(5) // wraps from 4 to 0
	if vp.Cursor != 0 {
		t.Errorf("expected Cursor 0, got %d", vp.Cursor)
	}
	if vp.Offset != 0 {
		t.Errorf("expected Offset 0 after wrap to top, got %d", vp.Offset)
	}
}

func TestMoveDown_ShiftsViewportWhenCursorLeavesWindow(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	for i := 0; i < 10; i++ {
		vp.MoveDown(20)
	}
	if vp.Cursor != 10 {
		t.Fatalf("expected Cursor 10, got %d", vp.Cursor)
	}
	if vp.Cursor < vp.Offset || vp.Cursor >= vp.Offset+vp.Height {
		t.Errorf("cursor %d outside visible window [%d, %d)", vp.Cursor, vp.Offset, vp.Offset+vp.Height)
	}
}

func TestMoveDown_NoopOnEmptyList(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	vp.MoveDown(0)
	if vp.Cursor != 0 || vp.Offset != 0 {
		t.Errorf("expected no change on empty list, got Cursor=%d Offset=%d", vp.Cursor, vp.Offset)
	}
}

// --- MoveUp tests ---

func TestMoveUp_RetreatsCursor(t *testing.T) {
	vp := viewport.Viewport{Cursor: 2, Offset: 0, Height: 10}
	vp.MoveUp(5)
	if vp.Cursor != 1 {
		t.Errorf("expected Cursor 1, got %d", vp.Cursor)
	}
}

func TestMoveUp_WrapsToBottom(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	vp.MoveUp(15) // wraps from 0 to 14
	if vp.Cursor != 14 {
		t.Fatalf("expected Cursor 14, got %d", vp.Cursor)
	}
	if vp.Cursor < vp.Offset || vp.Cursor >= vp.Offset+vp.Height {
		t.Errorf("cursor %d outside visible window [%d, %d)", vp.Cursor, vp.Offset, vp.Offset+vp.Height)
	}
}

func TestMoveUp_ShiftsViewportWhenCursorLeavesWindow(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	for i := 0; i < 15; i++ {
		vp.MoveDown(20)
	}
	offsetAtBottom := vp.Offset
	for i := 0; i < 10; i++ {
		vp.MoveUp(20)
	}
	if vp.Offset >= offsetAtBottom {
		t.Errorf("expected Offset to decrease; was %d, now %d", offsetAtBottom, vp.Offset)
	}
	if vp.Cursor < vp.Offset || vp.Cursor >= vp.Offset+vp.Height {
		t.Errorf("cursor %d outside visible window [%d, %d)", vp.Cursor, vp.Offset, vp.Offset+vp.Height)
	}
}

func TestMoveUp_NoopOnEmptyList(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	vp.MoveUp(0)
	if vp.Cursor != 0 || vp.Offset != 0 {
		t.Errorf("expected no change on empty list, got Cursor=%d Offset=%d", vp.Cursor, vp.Offset)
	}
}

// --- RenderList tests ---

func TestRenderList_ShowsCursorIndicator(t *testing.T) {
	vp := viewport.Viewport{Cursor: 1, Offset: 0, Height: 10}
	var sb strings.Builder
	vp.RenderList(&sb, []string{"alpha", "beta", "gamma"})
	out := sb.String()

	if !strings.Contains(out, "  > beta") {
		t.Errorf("expected cursor on 'beta'\n%s", out)
	}
	if strings.Contains(out, "  > alpha") || strings.Contains(out, "  > gamma") {
		t.Errorf("unexpected cursor indicator on non-selected item\n%s", out)
	}
}

func TestRenderList_OnlyRendersVisibleItems(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("item%02d", i)
	}
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	var sb strings.Builder
	vp.RenderList(&sb, items)
	out := sb.String()

	if !strings.Contains(out, "item00") {
		t.Errorf("expected item00 in output")
	}
	if strings.Contains(out, "item10") {
		t.Errorf("did not expect item10 in output (height=10)")
	}
}

func TestRenderList_ShowsScrollHints(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("item%02d", i)
	}
	vp := viewport.Viewport{Cursor: 10, Offset: 5, Height: 10}
	var sb strings.Builder
	vp.RenderList(&sb, items)
	out := sb.String()

	if !strings.Contains(out, "more above") {
		t.Errorf("expected 'more above' hint\n%s", out)
	}
	if !strings.Contains(out, "more below") {
		t.Errorf("expected 'more below' hint\n%s", out)
	}
}

func TestRenderList_NoScrollHintsWhenAllFit(t *testing.T) {
	vp := viewport.Viewport{Cursor: 0, Offset: 0, Height: 10}
	var sb strings.Builder
	vp.RenderList(&sb, []string{"a", "b", "c"})
	out := sb.String()

	if strings.Contains(out, "more above") || strings.Contains(out, "more below") {
		t.Errorf("expected no scroll hints\n%s", out)
	}
}
