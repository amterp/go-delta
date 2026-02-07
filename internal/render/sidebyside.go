package render

import (
	"fmt"
	"strings"

	"github.com/amterp/go-delta/internal/align"
)

// RenderSideBySide produces a two-panel diff. Left panel shows old text,
// right panel shows new text, separated by " │ ".
func RenderSideBySide(hunks []align.AnnotatedHunk, s Styles, termWidth int) string {
	if len(hunks) == 0 {
		return ""
	}

	panelWidth := (termWidth - 3) / 2 // 3 for " │ "
	if panelWidth < 10 {
		panelWidth = 10
	}

	maxOld, maxNew := maxLineNumbers(hunks)
	oldNumWidth := digitCount(maxOld)
	newNumWidth := digitCount(maxNew)

	var b strings.Builder
	sep := " │ "

	for i, h := range hunks {
		if i > 0 {
			b.WriteString("\n")
		}

		if h.Skipped > 0 {
			line := fmt.Sprintf("~~~ %d lines skipped ~~~", h.Skipped)
			centered := centerPad(s.Separator(line), termWidth)
			b.WriteString(centered)
			b.WriteString("\n\n")
		}

		rows := walkHunk(h)

		oldNum := h.OldStart
		newNum := h.NewStart

		for _, row := range rows {
			var left, right string

			switch {
			case row.IsContext:
				left = sbsPanel(s.LineNum(formatLineNum(oldNum, oldNumWidth)),
					s.Plain("  "+row.Left.Content), panelWidth)
				right = sbsPanel(s.LineNum(formatLineNum(newNum, newNumWidth)),
					s.Plain("  "+row.Right.Content), panelWidth)
				oldNum++
				newNum++

			case row.IsPaired:
				leftContent := RenderAnnotatedLine(row.Pair.Alignment.Old, s.Removed, s.RemovedEmph)
				left = sbsPanel(s.LineNum(formatLineNum(oldNum, oldNumWidth)),
					s.Removed("- ")+leftContent, panelWidth)
				rightContent := RenderAnnotatedLine(row.Pair.Alignment.New, s.Added, s.AddedEmph)
				right = sbsPanel(s.LineNum(formatLineNum(newNum, newNumWidth)),
					s.Added("+ ")+rightContent, panelWidth)
				oldNum++
				newNum++

			case row.Left != nil:
				left = sbsPanel(s.LineNum(formatLineNum(oldNum, oldNumWidth)),
					s.Removed("- "+row.Left.Content), panelWidth)
				right = sbsEmptyPanel(s, newNumWidth, panelWidth)
				oldNum++

			case row.Right != nil:
				left = sbsEmptyPanel(s, oldNumWidth, panelWidth)
				right = sbsPanel(s.LineNum(formatLineNum(newNum, newNumWidth)),
					s.Added("+ "+row.Right.Content), panelWidth)
				newNum++
			}

			b.WriteString(left + sep + right + "\n")
		}
	}

	return b.String()
}

// sbsPanel formats one side-by-side panel: "NN │ content", padded or
// truncated to fit panelWidth. Truncated lines end with "…" so the
// user knows content was cut off.
func sbsPanel(numStr, content string, panelWidth int) string {
	inner := numStr + " │ " + content
	vl := visibleWidth(inner)
	if vl <= panelWidth {
		return inner + strings.Repeat(" ", panelWidth-vl)
	}
	// Content exceeds panel width - truncate and add indicator.
	truncated := truncateToWidth(inner, panelWidth-1)
	return truncated + "…"
}

// sbsEmptyPanel creates a blank panel with a dimmed "~" placeholder.
func sbsEmptyPanel(s Styles, numWidth, panelWidth int) string {
	inner := blankLineNum(numWidth) + " │ " + s.Separator("~")
	vl := visibleWidth(inner)
	if vl >= panelWidth {
		return inner
	}
	return inner + strings.Repeat(" ", panelWidth-vl)
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
