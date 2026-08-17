package tui

// viewport.go holds the pure list-windowing math for the scrolling viewport.
// Keeping it side-effect free makes it directly table-testable without a TTY.

// scrollMargin is the fixed scroll-off: the window starts scrolling this many
// rows before the cursor reaches the top/bottom edge. It is capped for short
// terminals in window().
const scrollMargin = 2

// window computes the visible slice for a list given the cursor and the previous
// offset. It returns the (possibly updated) offset and the [first, last) slice
// bounds into a list of n rows.
//
// Rules:
//   - visible = height - chrome. If height <= 0 this is the "no constraint"
//     sentinel: render all rows (offset 0, first 0, last n).
//   - visible is clamped to at least 1.
//   - margin is scrollMargin capped to (visible-1)/2 so it fits short terminals.
//   - the scroll rule slides offset to keep the cursor within the margin, then
//     offset is clamped to [0, max(0, n-visible)].
//   - first = offset, last = min(offset+visible, n).
func window(cursor, offset, height, chrome, n int) (newOffset, first, last int) {
	// Sentinel: unknown terminal height -> render everything.
	if height <= 0 {
		return 0, 0, n
	}

	visible := height - chrome
	if visible < 1 {
		visible = 1
	}

	margin := scrollMargin
	if cap := (visible - 1) / 2; margin > cap {
		margin = cap
	}

	// Scroll rule: slide the window to keep the cursor within the margin.
	if cursor-margin < offset {
		offset = cursor - margin
	}
	if cursor+margin >= offset+visible {
		offset = cursor + margin - visible + 1
	}

	// Clamp offset so the window never runs off either end of the list.
	maxOffset := n - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	first = offset
	last = offset + visible
	if last > n {
		last = n
	}
	return offset, first, last
}

// chrome returns the number of non-list lines View writes for the current state:
// header (title + blank = 2) + optional filter line + optional status line +
// position indicator (1) + footer (1). It must match View exactly so the last
// list row is never clipped.
func (m *Model) chrome() int {
	c := 2 // title + blank line
	if m.filtering || m.filter.Value() != "" {
		c++ // filter input line
	}
	if m.facets.status() != "" {
		c++ // facet-status line
	}
	if m.status != "" {
		c++ // status line
	}
	c++ // position indicator line
	c++ // footer line
	return c
}

// visibleRows returns the number of list rows that fit given the current height
// and chrome. It returns the sentinel 0 when the height is unknown (height <= 0)
// so callers can treat that as "no constraint". Update and View both use this so
// paging keys agree with rendering.
func (m *Model) visibleRows() int {
	if m.height <= 0 {
		return 0
	}
	v := m.height - m.chrome()
	if v < 1 {
		v = 1
	}
	return v
}

// applyWindow recomputes and stores m.offset for the current cursor and list
// length, returning the [first, last) slice bounds. View calls this before
// rendering each list.
func (m *Model) applyWindow(n int) (first, last int) {
	m.offset, first, last = window(m.cursor, m.offset, m.height, m.chrome(), n)
	return first, last
}
