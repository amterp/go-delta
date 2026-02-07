package render

import (
	"fmt"
	"strings"

	"github.com/amterp/go-delta/internal/align"
	"github.com/amterp/go-delta/internal/diff"
	"github.com/mattn/go-runewidth"
)

// formatLineNum formats a line number right-justified to the given width.
func formatLineNum(n int, width int) string {
	return fmt.Sprintf("%*d", width, n)
}

// blankLineNum returns a blank string of the given width.
func blankLineNum(width int) string {
	return strings.Repeat(" ", width)
}

// digitCount returns the number of digits in a non-negative integer.
func digitCount(n int) int {
	if n == 0 {
		return 1
	}
	count := 0
	for n > 0 {
		count++
		n /= 10
	}
	return count
}

// --- Gutter formatting ---

// gutterInline formats the dual-number gutter for inline mode.
// For context lines: "NN MM │ "
// For delete lines:  "NN    │ "
// For insert lines:  "   MM │ "
func gutterInline(kind diff.OpKind, oldNum, newNum, oldWidth, newWidth int, s Styles) string {
	sep := s.LineNum("│")
	switch kind {
	case diff.OpEqual:
		return fmt.Sprintf("%s %s %s ",
			s.LineNum(formatLineNum(oldNum, oldWidth)),
			s.LineNum(formatLineNum(newNum, newWidth)),
			sep)
	case diff.OpDelete:
		return fmt.Sprintf("%s %s %s ",
			s.LineNum(formatLineNum(oldNum, oldWidth)),
			blankLineNum(newWidth),
			sep)
	case diff.OpInsert:
		return fmt.Sprintf("%s %s %s ",
			blankLineNum(oldWidth),
			s.LineNum(formatLineNum(newNum, newWidth)),
			sep)
	}
	return ""
}

// --- Hunk walking ---

// hunkRow represents a single output row produced by walking a hunk.
// Exactly one of the following patterns holds:
//   - IsContext: both Left and Right are set (same content)
//   - IsPaired: both Left and Right are set, Pair has alignment data
//   - Left only: unpaired delete
//   - Right only: unpaired insert
type hunkRow struct {
	IsContext bool
	IsPaired  bool
	Left      *diff.Line       // the old/removed line, or nil
	Right     *diff.Line       // the new/added line, or nil
	Pair      *align.LinePair  // non-nil only for paired rows
}

// walkHunk flattens an annotated hunk into a sequence of rows suitable
// for either inline or side-by-side rendering. This eliminates the
// fragile manual index-tracking that side-by-side previously required.
func walkHunk(h align.AnnotatedHunk) []hunkRow {
	// Build pair lookups
	oldPairs := make(map[int]*align.LinePair)
	newPairs := make(map[int]bool)
	for i := range h.Pairs {
		p := &h.Pairs[i]
		oldPairs[p.OldIdx] = p
		newPairs[p.NewIdx] = true
	}

	var rows []hunkRow
	for li := 0; li < len(h.Lines); li++ {
		line := h.Lines[li]
		switch line.Kind {
		case diff.OpEqual:
			rows = append(rows, hunkRow{
				IsContext: true,
				Left:     &h.Lines[li],
				Right:    &h.Lines[li],
			})

		case diff.OpDelete:
			if pair, ok := oldPairs[li]; ok {
				rows = append(rows, hunkRow{
					IsPaired: true,
					Left:     &h.Lines[li],
					Right:    &h.Lines[pair.NewIdx],
					Pair:     pair,
				})
			} else {
				rows = append(rows, hunkRow{
					Left: &h.Lines[li],
				})
			}

		case diff.OpInsert:
			if newPairs[li] {
				continue // already emitted by the paired delete
			}
			rows = append(rows, hunkRow{
				Right: &h.Lines[li],
			})
		}
	}
	return rows
}

// --- Styled line rendering ---

// RenderAnnotatedLine reconstructs a line from aligned tokens, applying
// emphasis styling to changed tokens and base styling to matched tokens.
// Consecutive tokens with the same treatment are grouped to minimize
// style transitions (and ANSI escape overhead).
func RenderAnnotatedLine(tokens []align.AlignedToken, baseStyle, emphStyle func(string) string) string {
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	var buf strings.Builder
	currentEmph := false

	flush := func() {
		text := buf.String()
		if text == "" {
			return
		}
		if currentEmph {
			b.WriteString(emphStyle(text))
		} else {
			b.WriteString(baseStyle(text))
		}
		buf.Reset()
	}

	for _, at := range tokens {
		isEmph := at.Op != align.AlignMatch
		if isEmph != currentEmph && buf.Len() > 0 {
			flush()
		}
		currentEmph = isEmph
		buf.WriteString(at.Token.Text)
	}
	flush()

	return b.String()
}

// --- ANSI-aware string measurement ---

// visibleWidth returns the display width of a string, accounting for
// ANSI escape sequences (skipped) and wide Unicode characters (CJK,
// emoji) which occupy two terminal columns.
func visibleWidth(s string) int {
	width := 0
	state := ansiNone
	for _, r := range s {
		prev := state
		state = ansiNext(state, r)
		if state != ansiNone || prev != ansiNone {
			continue // inside escape, or this rune terminated one
		}
		width += runewidth.RuneWidth(r)
	}
	return width
}

// truncateToWidth truncates a string to fit within maxWidth visible
// columns. ANSI escape sequences are preserved (they don't consume
// visible space). Wide characters that would exceed the limit are
// replaced with a space to maintain exact width.
func truncateToWidth(s string, maxWidth int) string {
	var b strings.Builder
	width := 0
	state := ansiNone

	for _, r := range s {
		prevState := state
		state = ansiNext(state, r)

		if state != ansiNone {
			// Inside an escape sequence - always emit
			b.WriteRune(r)
			continue
		}
		if prevState != ansiNone {
			// This rune terminated an escape - emit it but don't count width
			b.WriteRune(r)
			continue
		}

		rw := runewidth.RuneWidth(r)
		if width+rw > maxWidth {
			if width < maxWidth {
				b.WriteRune(' ')
			}
			break
		}
		b.WriteRune(r)
		width += rw
	}

	return b.String()
}

// ANSI escape sequence state machine.
// Handles ESC (0x1B) followed by:
//   - '[' -> CSI: parameters (0x30-0x3F), intermediates (0x20-0x2F), final byte (0x40-0x7E)
//   - ']' -> OSC: terminated by BEL (0x07) or ST (ESC \)
//   - 'P', '_', '^', 'X' -> DCS/APC/PM/SOS: terminated by ST (ESC \)
//   - Other single-character sequences (e.g., ESC 7, ESC 8)
type ansiState int

const (
	ansiNone   ansiState = iota // not in an escape
	ansiEsc                     // saw ESC, waiting for next char
	ansiCSI                     // inside CSI sequence (ESC [), waiting for final byte
	ansiOSC                     // inside OSC (ESC ]), waiting for BEL or ST
	ansiString                  // inside DCS/APC/PM/SOS, waiting for ST (ESC \)
	ansiStringEsc               // inside string sequence, saw ESC (possible ST)
)

func ansiNext(state ansiState, r rune) ansiState {
	switch state {
	case ansiNone:
		if r == '\x1b' {
			return ansiEsc
		}
		return ansiNone
	case ansiEsc:
		switch r {
		case '[':
			return ansiCSI
		case ']':
			return ansiOSC
		case 'P', '_', '^', 'X':
			return ansiString
		default:
			return ansiNone // single-char escape (e.g., ESC 7)
		}
	case ansiCSI:
		// CSI parameters are 0x30-0x3F, intermediates 0x20-0x2F,
		// final byte is 0x40-0x7E
		if r >= 0x40 && r <= 0x7E {
			return ansiNone // sequence complete
		}
		return ansiCSI // still in parameters/intermediates
	case ansiOSC:
		if r == '\x07' {
			return ansiNone // BEL terminates OSC
		}
		if r == '\x1b' {
			return ansiStringEsc // possible ST (ESC \)
		}
		return ansiOSC
	case ansiString:
		if r == '\x1b' {
			return ansiStringEsc // possible ST (ESC \)
		}
		return ansiString
	case ansiStringEsc:
		if r == '\\' {
			return ansiNone // ST terminates the sequence
		}
		// ESC followed by non-backslash inside a string sequence.
		// Malformed, but treat the ESC as starting a new escape to
		// avoid consuming visible text.
		return ansiNext(ansiEsc, r)
	}
	return ansiNone
}
