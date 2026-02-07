package render

import (
	"fmt"
	"strings"

	"github.com/amterp/go-delta/internal/align"
)

// sbsItem is a single output element: either a separator line or a
// left/right panel pair.
type sbsItem struct {
	separator string // non-empty for hunk separator lines
	left      string // left panel content (no trailing padding)
	right     string // right panel content
}

// RenderSideBySide produces a two-panel diff. Left panel shows old text,
// right panel shows new text, separated by " │ ".
//
// Panels are sized to fit their content rather than always filling half
// the terminal width. termWidth is used only as the truncation ceiling.
func RenderSideBySide(hunks []align.AnnotatedHunk, s Styles, termWidth int) string {
	if len(hunks) == 0 {
		return ""
	}

	maxPanelWidth := (termWidth - 3) / 2 // 3 for " │ "
	if maxPanelWidth < 10 {
		maxPanelWidth = 10
	}

	maxOld, maxNew := maxLineNumbers(hunks)
	oldNumWidth := digitCount(maxOld)
	newNumWidth := digitCount(maxNew)

	// First pass: build all panel content (truncated, not padded) and
	// track the actual max widths needed.
	var items []sbsItem
	maxLeftVW := 0
	maxRightVW := 0

	for i, h := range hunks {
		if i > 0 {
			items = append(items, sbsItem{separator: "\n"})
		}

		if h.Skipped > 0 {
			noun := "lines"
			if h.Skipped == 1 {
				noun = "line"
			}
			line := fmt.Sprintf("~~~ %d %s skipped ~~~", h.Skipped, noun)
			items = append(items, sbsItem{separator: s.Separator(line)})
		}

		rows := walkHunk(h)

		oldNum := h.OldStart
		newNum := h.NewStart

		for _, row := range rows {
			var left, right string

			switch {
			case row.IsContext:
				left = sbsPanelContent(s.LineNum(formatLineNum(oldNum, oldNumWidth)),
					s.Plain("  "+row.Left.Content), maxPanelWidth)
				right = sbsPanelContent(s.LineNum(formatLineNum(newNum, newNumWidth)),
					s.Plain("  "+row.Right.Content), maxPanelWidth)
				oldNum++
				newNum++

			case row.IsPaired:
				leftContent := RenderAnnotatedLine(row.Pair.Alignment.Old, s.Removed, s.RemovedEmph)
				left = sbsPanelContent(s.LineNum(formatLineNum(oldNum, oldNumWidth)),
					s.Removed("- ")+leftContent, maxPanelWidth)
				rightContent := RenderAnnotatedLine(row.Pair.Alignment.New, s.Added, s.AddedEmph)
				right = sbsPanelContent(s.LineNum(formatLineNum(newNum, newNumWidth)),
					s.Added("+ ")+rightContent, maxPanelWidth)
				oldNum++
				newNum++

			case row.Left != nil:
				left = sbsPanelContent(s.LineNum(formatLineNum(oldNum, oldNumWidth)),
					s.Removed("- "+row.Left.Content), maxPanelWidth)
				right = sbsEmptyContent(s, newNumWidth)
				oldNum++

			case row.Right != nil:
				left = sbsEmptyContent(s, oldNumWidth)
				right = sbsPanelContent(s.LineNum(formatLineNum(newNum, newNumWidth)),
					s.Added("+ "+row.Right.Content), maxPanelWidth)
				newNum++
			}

			if vw := visibleWidth(left); vw > maxLeftVW {
				maxLeftVW = vw
			}
			if vw := visibleWidth(right); vw > maxRightVW {
				maxRightVW = vw
			}
			items = append(items, sbsItem{left: left, right: right})
		}
	}

	// Second pass: pad left panels to maxLeftVW, join with separator.
	var b strings.Builder
	sep := " │ "
	totalContentWidth := maxLeftVW + 3 + maxRightVW

	for _, item := range items {
		if item.separator != "" {
			if item.separator == "\n" {
				b.WriteString("\n")
			} else {
				centered := centerPad(item.separator, totalContentWidth)
				b.WriteString(centered)
				b.WriteString("\n\n")
			}
			continue
		}

		padded := item.left + strings.Repeat(" ", maxLeftVW-visibleWidth(item.left))
		b.WriteString(padded + sep + item.right + "\n")
	}

	return b.String()
}

// sbsPanelContent formats one panel's content: "NN │ content", truncated
// to maxWidth if needed. No trailing padding - that's applied in the
// second pass.
func sbsPanelContent(numStr, content string, maxWidth int) string {
	inner := numStr + " │ " + content
	vl := visibleWidth(inner)
	if vl <= maxWidth {
		return inner
	}
	return truncateToWidth(inner, maxWidth-1) + "…"
}

// sbsEmptyContent creates a blank panel with a dimmed "~" placeholder.
// No trailing padding.
func sbsEmptyContent(s Styles, numWidth int) string {
	return blankLineNum(numWidth) + " │ " + s.Separator("~")
}

// centerPad centers text by prepending spaces (approximate).
func centerPad(s string, width int) string {
	vl := visibleWidth(s)
	if vl >= width {
		return s
	}
	pad := (width - vl) / 2
	return strings.Repeat(" ", pad) + s
}
