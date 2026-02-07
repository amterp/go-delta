package render

import (
	"fmt"
	"strings"

	"github.com/amterp/go-delta/internal/align"
	"github.com/amterp/go-delta/internal/diff"
)

// RenderInline produces an inline (unified-style) diff string from
// annotated hunks. Paired lines get within-line emphasis; unpaired
// lines are rendered entirely in their base color.
func RenderInline(hunks []align.AnnotatedHunk, s Styles) string {
	if len(hunks) == 0 {
		return ""
	}

	maxOld, maxNew := maxLineNumbers(hunks)
	oldWidth := digitCount(maxOld)
	newWidth := digitCount(maxNew)

	var b strings.Builder

	for i, h := range hunks {
		if i > 0 {
			b.WriteString("\n")
		}

		if h.Skipped > 0 {
			noun := "lines"
			if h.Skipped == 1 {
				noun = "line"
			}
			sep := fmt.Sprintf("~~~ %d %s skipped ~~~", h.Skipped, noun)
			b.WriteString(s.Separator(sep))
			b.WriteString("\n\n")
		}

		rows := walkHunk(h)

		oldNum := h.OldStart
		newNum := h.NewStart

		for _, row := range rows {
			switch {
			case row.IsContext:
				gutter := gutterInline(diff.OpEqual, oldNum, newNum, oldWidth, newWidth, s)
				b.WriteString(gutter + "  " + row.Left.Content + "\n")
				oldNum++
				newNum++

			case row.IsPaired:
				// Build both gutters before incrementing either counter
				delGutter := gutterInline(diff.OpDelete, oldNum, newNum, oldWidth, newWidth, s)
				insGutter := gutterInline(diff.OpInsert, oldNum, newNum, oldWidth, newWidth, s)

				annotated := RenderAnnotatedLine(row.Pair.Alignment.Old, s.Removed, s.RemovedEmph)
				b.WriteString(delGutter + s.Removed("- ") + annotated + "\n")

				annotated = RenderAnnotatedLine(row.Pair.Alignment.New, s.Added, s.AddedEmph)
				b.WriteString(insGutter + s.Added("+ ") + annotated + "\n")

				oldNum++
				newNum++

			case row.Left != nil:
				gutter := gutterInline(diff.OpDelete, oldNum, newNum, oldWidth, newWidth, s)
				b.WriteString(gutter + s.Removed(fmt.Sprintf("- %s", row.Left.Content)) + "\n")
				oldNum++

			case row.Right != nil:
				gutter := gutterInline(diff.OpInsert, oldNum, newNum, oldWidth, newWidth, s)
				b.WriteString(gutter + s.Added(fmt.Sprintf("+ %s", row.Right.Content)) + "\n")
				newNum++
			}
		}
	}

	return b.String()
}

// maxLineNumbers computes the highest old and new line numbers that
// will be displayed across all hunks, so gutter widths can be fixed.
func maxLineNumbers(hunks []align.AnnotatedHunk) (maxOld, maxNew int) {
	for _, h := range hunks {
		oldNum := h.OldStart
		newNum := h.NewStart
		for _, line := range h.Lines {
			switch line.Kind {
			case diff.OpEqual:
				if oldNum > maxOld {
					maxOld = oldNum
				}
				if newNum > maxNew {
					maxNew = newNum
				}
				oldNum++
				newNum++
			case diff.OpDelete:
				if oldNum > maxOld {
					maxOld = oldNum
				}
				oldNum++
			case diff.OpInsert:
				if newNum > maxNew {
					maxNew = newNum
				}
				newNum++
			}
		}
	}
	return
}
