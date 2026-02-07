package align

import "github.com/amterp/go-delta/internal/diff"

// DistanceThreshold controls the maximum normalized edit distance for
// two lines to be considered a pair. Must be in [0.0, 1.0]. Lines at
// or above this threshold are treated as unrelated (fully added/removed,
// no within-line emphasis).
//
// 0.6 strikes a balance: low enough to avoid misleading highlights when
// lines are completely rewritten, high enough to catch most single-token
// edits within a line. This matches Delta's approach.
const DistanceThreshold = 0.6

// maxAlignTokens is the maximum token count (per line) for NW alignment.
// The NW algorithm allocates an O(N*M) matrix, so very long lines (e.g.
// minified JS) could exhaust memory. Lines exceeding this are left
// unpaired rather than risking a multi-GB allocation.
const maxAlignTokens = 500

// LinePair records a pairing between a removed and added line in a hunk,
// along with their token-level alignment.
type LinePair struct {
	OldIdx    int       // index into Hunk.Lines of the removed line
	NewIdx    int       // index into Hunk.Lines of the added line
	Alignment Alignment // token-level alignment result
}

// AnnotatedHunk wraps a diff.Hunk with line-pairing information.
type AnnotatedHunk struct {
	diff.Hunk
	Pairs []LinePair
}

// AnnotateHunks performs greedy forward-search line pairing on each hunk.
// For each removed line, it scans forward through unpaired added lines,
// tokenizes both, runs NW alignment, and pairs them if the distance is
// below DistanceThreshold.
func AnnotateHunks(hunks []diff.Hunk) []AnnotatedHunk {
	result := make([]AnnotatedHunk, len(hunks))
	for i, h := range hunks {
		result[i] = annotateHunk(h)
	}
	return result
}

func annotateHunk(h diff.Hunk) AnnotatedHunk {
	ah := AnnotatedHunk{Hunk: h}

	// Track which added lines have been paired
	paired := make(map[int]bool)

	for i, line := range h.Lines {
		if line.Kind != diff.OpDelete {
			continue
		}

		oldTokens := Tokenize(line.Content)
		if len(oldTokens) > maxAlignTokens {
			continue // too many tokens, skip alignment
		}

		// Scan forward for an unpaired added line
		for j := i + 1; j < len(h.Lines); j++ {
			if h.Lines[j].Kind == diff.OpEqual {
				break // stop at context boundary
			}
			if h.Lines[j].Kind != diff.OpInsert {
				continue
			}
			if paired[j] {
				continue
			}

			newTokens := Tokenize(h.Lines[j].Content)
			if len(newTokens) > maxAlignTokens {
				continue // too many tokens, skip alignment
			}
			alignment := Align(oldTokens, newTokens)

			if alignment.Distance < DistanceThreshold {
				ah.Pairs = append(ah.Pairs, LinePair{
					OldIdx:    i,
					NewIdx:    j,
					Alignment: alignment,
				})
				paired[j] = true
				break // this removed line is now paired
			}
		}
	}

	return ah
}
