package render

// Styles holds formatter functions for each visual element.
// Each function takes a string and returns it styled (or as-is for
// no-color mode). This keeps the render package decoupled from any
// specific color library.
type Styles struct {
	Removed     func(string) string // removed line text
	Added       func(string) string // added line text
	RemovedEmph func(string) string // emphasized segment in removed line
	AddedEmph   func(string) string // emphasized segment in added line
	LineNum     func(string) string // line numbers in gutter
	Separator   func(string) string // hunk separator text
	Plain       func(string) string // default / context text
}

// NoColorStyles returns Styles where every formatter is the identity
// function. Useful for no-color mode and testing.
func NoColorStyles() Styles {
	id := func(s string) string { return s }
	return Styles{
		Removed:     id,
		Added:       id,
		RemovedEmph: id,
		AddedEmph:   id,
		LineNum:     id,
		Separator:   id,
		Plain:       id,
	}
}
