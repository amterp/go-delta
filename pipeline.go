package godelta

import (
	"os"

	"github.com/amterp/go-delta/internal/align"
	"github.com/amterp/go-delta/internal/diff"
	"github.com/amterp/go-delta/internal/render"
	"golang.org/x/term"
)

// runPipeline executes the three-stage diff pipeline:
// 1. Line-level diff (Myers) -> hunks
// 2. Within-line alignment (tokenize + NW + line pairing)
// 3. Rendering
func runPipeline(old, new string, cfg config, styles render.Styles) string {
	// Stage 1: line-level diff
	lines := diff.Diff(old, new)
	if lines == nil {
		return ""
	}
	hunks := diff.ComputeHunks(lines, cfg.contextLines)
	if len(hunks) == 0 {
		return ""
	}

	// Stage 2: within-line alignment
	annotated := align.AnnotateHunks(hunks)

	// Stage 3: rendering
	if cfg.sideBySide {
		width := cfg.width
		if width <= 0 {
			width = terminalWidth()
		}
		return render.RenderSideBySide(annotated, styles, width)
	}
	return render.RenderInline(annotated, styles)
}

// terminalWidth detects the terminal width, returning 0 if detection
// fails. A zero value tells the renderer to skip truncation.
func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 0
	}
	return w
}
