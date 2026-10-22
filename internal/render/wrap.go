package render

import (
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

func writeInlineLine(b *strings.Builder, line string, width, indent int) {
	if width <= 0 || visibleWidth(line) <= width {
		b.WriteString(line)
		b.WriteByte('\n')
		return
	}

	textWidth := max(1, width-indent)
	var styles wrapStyles
	for first := true; len(line) > 0; first = false {
		prefixWidth := 0
		if first {
			prefixWidth = indent
		} else {
			b.WriteString(strings.Repeat(" ", indent))
			b.WriteString(styles.hyperlink)
			b.WriteString(styles.sgr)
		}
		end := inlineBreak(line, textWidth, prefixWidth)
		part := line[:end]
		b.WriteString(part)
		styles.consume(part)
		if styles.sgr != "" {
			b.WriteString("\x1b[0m")
		}
		if styles.hyperlink != "" {
			b.WriteString("\x1b]8;;\x1b\\")
		}
		b.WriteByte('\n')
		line = line[end:]
	}
}

func inlineBreak(line string, width, prefixWidth int) int {
	columns := -prefixWidth
	lastSpace := 0
	state := ansiNone
	for i, r := range line {
		prev := state
		state = ansiNext(state, r)
		if state != ansiNone || prev != ansiNone {
			continue
		}
		rw := max(0, runewidth.RuneWidth(r))
		if rw > 0 && columns > 0 && columns+rw > width {
			if lastSpace > 0 {
				return lastSpace
			}
			return i
		}
		columns += rw
		if columns > 0 && unicode.IsSpace(r) {
			lastSpace = i + len(string(r))
		}
	}
	return len(line)
}

// A physical newline must close styles before the continuation's blank gutter.
// Replaying SGR sequences retains selective resets and extended color parameters.
type wrapStyles struct {
	sgr       string
	hyperlink string
}

func (s *wrapStyles) consume(text string) {
	state := ansiNone
	start := 0
	for i, r := range text {
		prev := state
		state = ansiNext(state, r)
		if prev == ansiNone && state == ansiEsc {
			start = i
		}
		if prev == ansiNone || state != ansiNone {
			continue
		}
		sequence := text[start : i+len(string(r))]
		switch {
		case strings.HasPrefix(sequence, "\x1b[") && strings.HasSuffix(sequence, "m"):
			if sequence == "\x1b[0m" || sequence == "\x1b[m" {
				s.sgr = ""
			} else {
				s.sgr += sequence
			}
		case strings.HasPrefix(sequence, "\x1b]8;"):
			body := strings.TrimSuffix(strings.TrimSuffix(sequence[4:], "\a"), "\x1b\\")
			_, uri, ok := strings.Cut(body, ";")
			if ok && uri != "" {
				s.hyperlink = sequence
			} else {
				s.hyperlink = ""
			}
		}
	}
}
