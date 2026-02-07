// Myers diff algorithm adapted for line-level diffing.
//
// Based on Eugene W. Myers' "An O(ND) Difference Algorithm and Its
// Variations" (1986). The algorithm from golang.org/x/tools informed
// this implementation; see LICENSE in this directory.
package diff

import "strings"

// Diff computes a line-level diff between two strings.
// It splits both inputs on newlines and returns a sequence of Lines
// classifying each line as equal, deleted, or inserted.
func Diff(old, new string) []Line {
	if old == new {
		return nil
	}

	oldLines := splitLines(old)
	newLines := splitLines(new)

	return diffLines(oldLines, newLines)
}

// splitLines splits s into lines. An empty string returns nil (zero
// lines), not a single empty-string element. This is correct because
// the caller (Diff) short-circuits the equal case before we get here,
// so "" only appears when the other side is non-empty, and nil lets
// diffLines emit pure inserts or pure deletes.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// diffLines runs the Myers algorithm on two slices of strings and
// returns the result as a flat sequence of Lines.
func diffLines(old, new []string) []Line {
	n := len(old)
	m := len(new)

	if n == 0 {
		lines := make([]Line, m)
		for i, l := range new {
			lines[i] = Line{Kind: OpInsert, Content: l}
		}
		return lines
	}
	if m == 0 {
		lines := make([]Line, n)
		for i, l := range old {
			lines[i] = Line{Kind: OpDelete, Content: l}
		}
		return lines
	}

	// Myers shortest-edit-script algorithm.
	// We compute the edit graph and trace back to find the path.
	max := n + m
	// v stores the x position of the furthest-reaching path for each
	// diagonal k. We use offset max so that v[k+max] maps to diagonal k.
	v := make([]int, 2*max+1)
	// trace stores a copy of v at each step d, used for backtracking.
	trace := make([][]int, 0, max)

	var found bool
	for d := 0; d <= max; d++ {
		snapshot := make([]int, len(v))
		copy(snapshot, v)
		trace = append(trace, snapshot)

		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[k-1+max] < v[k+1+max]) {
				x = v[k+1+max] // move down
			} else {
				x = v[k-1+max] + 1 // move right
			}
			y := x - k

			// follow diagonal (matching lines)
			for x < n && y < m && old[x] == new[y] {
				x++
				y++
			}

			v[k+max] = x

			if x >= n && y >= m {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	return backtrack(trace, old, new, max)
}

// backtrack reconstructs the edit sequence from the trace.
func backtrack(trace [][]int, old, new []string, max int) []Line {
	n := len(old)
	m := len(new)
	x, y := n, m

	// edits collected in reverse
	var edits []Line

	for d := len(trace) - 1; d >= 0; d-- {
		v := trace[d]
		k := x - y

		var prevK int
		if k == -d || (k != d && v[k-1+max] < v[k+1+max]) {
			prevK = k + 1 // came from above (insert)
		} else {
			prevK = k - 1 // came from left (delete)
		}

		prevX := v[prevK+max]
		prevY := prevX - prevK

		// diagonal moves (equal lines)
		for x > prevX && y > prevY {
			x--
			y--
			edits = append(edits, Line{Kind: OpEqual, Content: old[x]})
		}

		if d > 0 {
			if x == prevX {
				// vertical move: insertion
				y--
				edits = append(edits, Line{Kind: OpInsert, Content: new[y]})
			} else {
				// horizontal move: deletion
				x--
				edits = append(edits, Line{Kind: OpDelete, Content: old[x]})
			}
		}
	}

	// reverse to get forward order
	for i, j := 0, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}

	return edits
}
