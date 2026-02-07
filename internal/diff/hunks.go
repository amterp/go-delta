package diff

// ComputeHunks groups a flat sequence of diff Lines into Hunks, each
// containing a contiguous region of changes with surrounding context.
// contextLines controls how many equal lines appear around each change.
// Adjacent or overlapping context regions are merged into a single hunk.
func ComputeHunks(lines []Line, contextLines int) []Hunk {
	if len(lines) == 0 {
		return nil
	}

	// Find indices of all change lines (deletes/inserts).
	var changeIdxs []int
	for i, l := range lines {
		if l.Kind != OpEqual {
			changeIdxs = append(changeIdxs, i)
		}
	}
	if len(changeIdxs) == 0 {
		return nil
	}

	// Build ranges: each change gets a context window [start, end).
	// We then merge overlapping/adjacent windows into hunks.
	type window struct {
		start, end int
	}

	var windows []window
	for _, idx := range changeIdxs {
		start := idx - contextLines
		if start < 0 {
			start = 0
		}
		end := idx + contextLines + 1
		if end > len(lines) {
			end = len(lines)
		}

		if len(windows) > 0 && start <= windows[len(windows)-1].end {
			// merge with previous window
			if end > windows[len(windows)-1].end {
				windows[len(windows)-1].end = end
			}
		} else {
			windows = append(windows, window{start, end})
		}
	}

	// Convert windows to hunks with correct line numbers.
	hunks := make([]Hunk, len(windows))
	// We need to track the old/new line numbers as we walk through
	// all diff lines from the start to compute accurate starts.
	oldLine := 1 // 1-based
	newLine := 1

	windowIdx := 0
	for i, l := range lines {
		if windowIdx < len(windows) && i == windows[windowIdx].start {
			hunks[windowIdx].OldStart = oldLine
			hunks[windowIdx].NewStart = newLine
		}

		if windowIdx < len(windows) && i >= windows[windowIdx].start && i < windows[windowIdx].end {
			hunks[windowIdx].Lines = append(hunks[windowIdx].Lines, l)
		}

		// Advance line counters
		switch l.Kind {
		case OpEqual:
			oldLine++
			newLine++
		case OpDelete:
			oldLine++
		case OpInsert:
			newLine++
		}

		if windowIdx < len(windows) && i == windows[windowIdx].end-1 {
			windowIdx++
		}
	}

	// Compute Skipped counts: lines between end of previous visible
	// region and start of this one.
	prevEnd := 0
	for i := range hunks {
		if i == 0 {
			hunks[i].Skipped = windows[i].start
		} else {
			hunks[i].Skipped = windows[i].start - prevEnd
		}
		prevEnd = windows[i].end
	}

	return hunks
}
