package render

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/mattn/go-runewidth"
)

func writeInlineLine(b *strings.Builder, line string, width, indent int) {
	if width <= 0 {
		b.WriteString(line)
		b.WriteByte('\n')
		return
	}

	text := newInlineText(line)
	textWidth := max(1, width-indent)
	end := text.breakAt(0, textWidth, indent, indent)
	if end == len(text.plain) {
		b.WriteString(line)
		b.WriteByte('\n')
		return
	}

	var styles wrapStyles
	var sgr string
	start := 0
	rawStart := 0
	for first := true; start < len(text.plain); first = false {
		if !first {
			b.WriteString(strings.Repeat(" ", indent))
			b.WriteString(styles.hyperlink)
			b.WriteString(sgr)
			end = text.breakAt(start, textWidth, 0, indent)
		}
		rawEnd := text.offsets[end]
		part := line[rawStart:rawEnd]
		b.WriteString(part)
		styles.consume(part)
		sgr = styles.sgrSequence()
		if sgr != "" {
			b.WriteString("\x1b[0m")
		}
		if styles.hyperlink != "" {
			b.WriteString("\x1b]8;;\x1b\\")
		}
		b.WriteByte('\n')
		rawStart = rawEnd
		start = end
	}
}

type inlineText struct {
	plain   string
	offsets []int
}

// Segment visible text so ANSI sequences inside a grapheme do not create break points.
func newInlineText(line string) inlineText {
	var plain strings.Builder
	offsets := make([]int, 0, len(line)+1)
	state := ansiNone
	for i := 0; i < len(line); {
		r, size := utf8.DecodeRuneInString(line[i:])
		prev := state
		state = ansiNext(state, r)
		if state == ansiNone && prev == ansiNone {
			plain.WriteString(line[i : i+size])
			for j := 0; j < size; j++ {
				offsets = append(offsets, i+j)
			}
		}
		i += size
	}
	return inlineText{plain: plain.String(), offsets: append(offsets, len(line))}
}

func (t inlineText) breakAt(start, width, prefixWidth, indent int) int {
	columns := -prefixWidth
	lastSpace := 0
	g := graphemes.FromString(t.plain[start:])
	for g.Next() {
		cluster := g.Value()
		r, _ := utf8.DecodeRuneInString(cluster)
		rw := runewidth.StringWidth(cluster)
		if cluster == "\t" {
			rw = 8 - (columns+indent)%8
		}
		if rw > 0 && columns > 0 && columns+rw > width {
			if lastSpace > 0 {
				return lastSpace
			}
			return start + g.Start()
		}
		columns += rw
		if columns > 0 && unicode.IsSpace(r) {
			lastSpace = start + g.End()
		}
	}
	return len(t.plain)
}

// A physical newline must close styles before the continuation's blank gutter.
// Tracking current attributes bounds the ANSI sequences copied to continuations.
type wrapStyles struct {
	sgr       [108]string
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
			s.consumeSGR(sequence[2 : len(sequence)-1])
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

func (s *wrapStyles) consumeSGR(parameters string) {
	params := strings.Split(parameters, ";")
	for i := 0; i < len(params); i++ {
		param := params[i]
		number, _, _ := strings.Cut(param, ":")
		code := 0
		if number != "" {
			var err error
			code, err = strconv.Atoi(number)
			if err != nil {
				continue
			}
		}
		if (code == 38 || code == 48 || code == 58) && param == number && i+1 < len(params) {
			modeText, _, _ := strings.Cut(params[i+1], ":")
			mode, _ := strconv.Atoi(modeText)
			count := 0
			switch mode {
			case 5:
				count = 2
			case 2:
				count = 4
			}
			end := min(len(params), i+count+1)
			for j := i + 1; j < end; j++ {
				if strings.Contains(params[j], ":") {
					end = j + 1
					break
				}
			}
			param = strings.Join(params[i:end], ";")
			i = end - 1
		}
		key := code
		switch {
		case code == 0:
			s.sgr = [108]string{}
			continue
		case code == 22:
			s.sgr[1], s.sgr[2] = "", ""
			continue
		case code == 23:
			key, param = 3, ""
		case code == 24:
			key, param = 4, ""
		case code == 25:
			key, param = 5, ""
		case code == 27 || code == 28 || code == 29:
			key, param = code-20, ""
		case code == 39:
			key, param = 30, ""
		case code == 49:
			key, param = 40, ""
		case code == 50:
			key, param = 26, ""
		case code == 54:
			key, param = 51, ""
		case code == 55:
			key, param = 53, ""
		case code == 59:
			key, param = 58, ""
		case code == 65:
			clear(s.sgr[60:65])
			continue
		case code == 75:
			key, param = 73, ""
		case code == 20:
			key = 3
		case code == 21:
			key = 4
		case code == 6:
			key = 5
		case code >= 10 && code <= 19:
			key = 10
		case code >= 30 && code <= 38 || code >= 90 && code <= 97:
			key = 30
		case code >= 40 && code <= 48 || code >= 100 && code <= 107:
			key = 40
		case code == 52:
			key = 51
		case code == 74:
			key = 73
		}
		if key > 0 && key < len(s.sgr) {
			s.sgr[key] = param
		}
	}
}

func (s *wrapStyles) sgrSequence() string {
	var params []string
	for _, param := range s.sgr {
		if param != "" {
			params = append(params, param)
		}
	}
	if len(params) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(params, ";") + "m"
}
