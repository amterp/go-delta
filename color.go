package godelta

import (
	"os"

	"github.com/amterp/color"
	"github.com/amterp/go-delta/internal/render"
	"golang.org/x/term"
)

// newStyles builds the color objects for all visual elements.
// Callers should use buildStyles to get render.Styles closures.
type colorSet struct {
	removed     *color.Color
	added       *color.Color
	removedEmph *color.Color
	addedEmph   *color.Color
	lineNum     *color.Color
	separator   *color.Color
}

func newColorSet() colorSet {
	return colorSet{
		removed:     color.New(color.FgRed),
		added:       color.New(color.FgGreen),
		removedEmph: color.New(color.FgRed, color.ReverseVideo),
		addedEmph:   color.New(color.FgGreen, color.ReverseVideo),
		lineNum:     color.New(color.Faint),
		separator:   color.New(color.Faint),
	}
}

// buildStyles produces render.Styles from the color set, with colors
// enabled or disabled based on the resolved mode.
func buildStyles(useColor bool) render.Styles {
	if !useColor {
		return render.NoColorStyles()
	}

	cs := newColorSet()
	// Force color on each object so we don't depend on the global
	// NoColor flag (which auto-detects TTY). We've already made the
	// color decision ourselves in resolveColor.
	allColors := []*color.Color{
		cs.removed, cs.added, cs.removedEmph, cs.addedEmph,
		cs.lineNum, cs.separator,
	}
	for _, c := range allColors {
		c.EnableColor()
	}
	wrap := func(c *color.Color) func(string) string {
		return func(s string) string { return c.Sprint(s) }
	}
	return render.Styles{
		Removed:     wrap(cs.removed),
		Added:       wrap(cs.added),
		RemovedEmph: wrap(cs.removedEmph),
		AddedEmph:   wrap(cs.addedEmph),
		LineNum:     wrap(cs.lineNum),
		Separator:   wrap(cs.separator),
		Plain:       func(s string) string { return s },
	}
}

// resolveColor determines whether to use color output.
// Priority: explicit override > FORCE_COLOR > NO_COLOR > TTY detection.
func resolveColor(mode *bool) bool {
	if mode != nil {
		return *mode
	}
	if _, ok := os.LookupEnv("FORCE_COLOR"); ok {
		return true
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}
