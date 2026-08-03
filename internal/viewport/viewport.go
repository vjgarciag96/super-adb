// Package viewport provides a reusable scrolling viewport for Bubble Tea list components.
package viewport

import (
	"fmt"
	"strings"
)

// Viewport tracks cursor position, scroll offset, and visible height for a
// scrollable list. All fields are exported so callers can reset them directly
// (e.g. m.vp.Cursor = 0; m.vp.Offset = 0 on filter reset).
type Viewport struct {
	Cursor int // index of the highlighted item within the full list
	Offset int // index of the first visible item
	Height int // maximum number of list items to display at once
}

// SetHeight sets Height from a terminal height, subtracting reservedLines for
// chrome (header, footer, scroll hints). The result is clamped to minHeight.
func (vp *Viewport) SetHeight(terminalHeight, reservedLines, minHeight int) {
	vp.Height = terminalHeight - reservedLines
	if vp.Height < minHeight {
		vp.Height = minHeight
	}
}

// MoveUp moves the cursor one step up with wrapping and snaps the viewport.
// No-op when total is zero.
func (vp *Viewport) MoveUp(total int) {
	if total == 0 {
		return
	}
	vp.Cursor = (vp.Cursor - 1 + total) % total
	vp.Snap(total)
}

// MoveDown moves the cursor one step down with wrapping and snaps the viewport.
// No-op when total is zero.
func (vp *Viewport) MoveDown(total int) {
	if total == 0 {
		return
	}
	vp.Cursor = (vp.Cursor + 1) % total
	vp.Snap(total)
}

// Snap adjusts Offset so that Cursor stays within the visible window
// [Offset, Offset+Height). Call this after every Cursor change.
// total is the length of the currently visible item list.
func (vp *Viewport) Snap(total int) {
	if vp.Cursor < vp.Offset {
		vp.Offset = vp.Cursor
	} else if vp.Cursor >= vp.Offset+vp.Height {
		vp.Offset = vp.Cursor - vp.Height + 1
	}
	if vp.Offset < 0 {
		vp.Offset = 0
	}
}

// VisibleRange returns the [start, end) indices into the list that should be
// rendered. end is clamped to total. If Height <= 0 the full list is returned.
func (vp Viewport) VisibleRange(total int) (start, end int) {
	start = vp.Offset
	end = vp.Offset + vp.Height
	if vp.Height <= 0 || end > total {
		end = total
	}
	return start, end
}

// HiddenAbove returns the number of items scrolled past above the visible window.
func (vp Viewport) HiddenAbove() int { return vp.Offset }

// HiddenBelow returns the number of items below the visible window.
// total is the length of the currently visible item list.
// The result is always >= 0.
func (vp Viewport) HiddenBelow(total int) int {
	_, end := vp.VisibleRange(total)
	return total - end
}

// RenderList writes the visible portion of items into sb, with scroll hints
// above and below and a ">" cursor indicator on the highlighted item.
func (vp Viewport) RenderList(sb *strings.Builder, items []string) {
	start, end := vp.VisibleRange(len(items))

	if above := vp.HiddenAbove(); above > 0 {
		fmt.Fprintf(sb, "  (%d more above)\n", above)
	}
	for i := start; i < end; i++ {
		if i == vp.Cursor {
			fmt.Fprintf(sb, "  > %s\n", items[i])
		} else {
			fmt.Fprintf(sb, "    %s\n", items[i])
		}
	}
	if below := vp.HiddenBelow(len(items)); below > 0 {
		fmt.Fprintf(sb, "  (%d more below)\n", below)
	}
}
