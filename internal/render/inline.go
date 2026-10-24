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
func RenderInline(hunks []align.AnnotatedHunk, s Styles, width int) string {
	if len(hunks) == 0 {
		return ""
	}

	maxOld, maxNew := maxLineNumbers(hunks)
	oldWidth := digitCount(maxOld)
	newWidth := digitCount(maxNew)
	indent := oldWidth + newWidth + 5 + visibleWidth("│")

	var b strings.Builder
	var activeStyles wrapStyles

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
			separator := s.Separator(sep)
			b.WriteString(separator)
			if width > 0 {
				activeStyles.consume(separator)
			}
			b.WriteString("\n\n")
		}

		for _, row := range walkHunk(h) {
			switch {
			case row.IsContext:
				gutter := gutterInline(diff.OpEqual, row.OldNum, row.NewNum, oldWidth, newWidth, s)
				writeInlineLine(&b, gutter+"  "+row.Left.Content, width, indent, &activeStyles)

			case row.IsPaired:
				delGutter := gutterInline(diff.OpDelete, row.OldNum, row.NewNum, oldWidth, newWidth, s)
				insGutter := gutterInline(diff.OpInsert, row.OldNum, row.NewNum, oldWidth, newWidth, s)

				annotated := RenderAnnotatedLine(row.Pair.Alignment.Old, s.Removed, s.RemovedEmph)
				writeInlineLine(&b, delGutter+s.Removed("- ")+annotated, width, indent, &activeStyles)

				annotated = RenderAnnotatedLine(row.Pair.Alignment.New, s.Added, s.AddedEmph)
				writeInlineLine(&b, insGutter+s.Added("+ ")+annotated, width, indent, &activeStyles)

			case row.Left != nil:
				gutter := gutterInline(diff.OpDelete, row.OldNum, row.NewNum, oldWidth, newWidth, s)
				writeInlineLine(&b, gutter+s.Removed("- "+row.Left.Content), width, indent, &activeStyles)

			case row.Right != nil:
				gutter := gutterInline(diff.OpInsert, row.OldNum, row.NewNum, oldWidth, newWidth, s)
				writeInlineLine(&b, gutter+s.Added("+ "+row.Right.Content), width, indent, &activeStyles)
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
