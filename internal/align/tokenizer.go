package align

import (
	"unicode"
	"unicode/utf8"
)

// Token represents a segment of a line with its position.
type Token struct {
	Text  string
	Start int // byte offset in original line
	End   int // byte offset past last byte
}

// Tokenize splits a line into tokens suitable for word-level diffing.
// Rules:
//   - Runs of word characters (\w: letters, digits, underscore) form tokens
//   - Each non-word, non-space character is its own token
//   - Each whitespace character is its own token (for precise NW alignment)
//
// The tokenization is lossless: joining all token texts reproduces the
// original line exactly.
func Tokenize(line string) []Token {
	if line == "" {
		return nil
	}

	var tokens []Token
	runes := []rune(line)
	i := 0

	for i < len(runes) {
		start := byteOffset(runes, i)
		r := runes[i]

		switch {
		case isWordChar(r):
			// consume word run
			j := i + 1
			for j < len(runes) && isWordChar(runes[j]) {
				j++
			}
			end := byteOffset(runes, j)
			tokens = append(tokens, Token{
				Text:  line[start:end],
				Start: start,
				End:   end,
			})
			i = j

		case unicode.IsSpace(r):
			// individual whitespace characters so NW can align
			// runs that differ by only a few characters
			end := byteOffset(runes, i+1)
			tokens = append(tokens, Token{
				Text:  line[start:end],
				Start: start,
				End:   end,
			})
			i++

		default:
			// single punctuation/operator character
			end := byteOffset(runes, i+1)
			tokens = append(tokens, Token{
				Text:  line[start:end],
				Start: start,
				End:   end,
			})
			i++
		}
	}

	return tokens
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// byteOffset returns the byte offset of the i-th rune in a rune slice,
// relative to the original string.
func byteOffset(runes []rune, i int) int {
	offset := 0
	for j := 0; j < i; j++ {
		offset += utf8.RuneLen(runes[j])
	}
	return offset
}
