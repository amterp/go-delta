package godelta

// Option configures the behavior of DiffWith.
type Option func(*config)

// Layout controls how the diff is rendered.
type Layout int

const (
	// LayoutInline renders removals and additions on separate lines.
	LayoutInline Layout = iota
	// LayoutSideBySide renders old and new text in side-by-side panels,
	// truncating lines that exceed the terminal width.
	LayoutSideBySide
	// LayoutPreferSideBySide uses side-by-side when the content fits
	// within the terminal width, otherwise falls back to inline to
	// avoid truncating content. In non-TTY environments (where width
	// cannot be detected), uses side-by-side since there is no
	// truncation.
	LayoutPreferSideBySide
)

type config struct {
	contextLines int
	layout       Layout
	colorMode    *bool // nil = auto-detect
	width        int   // 0 = auto-detect terminal width
}

func defaultConfig() config {
	return config{
		contextLines: 3,
	}
}

// WithContextLines sets the number of unchanged lines shown around
// each change. Default is 3. Clamped to [0, 100000].
func WithContextLines(n int) Option {
	return func(c *config) {
		if n < 0 {
			n = 0
		}
		if n > 100000 {
			n = 100000
		}
		c.contextLines = n
	}
}

// WithLayout sets the diff layout mode. Default is LayoutInline.
func WithLayout(mode Layout) Option {
	return func(c *config) {
		c.layout = mode
	}
}

// WithColor forces color on or off, overriding auto-detection.
func WithColor(on bool) Option {
	return func(c *config) {
		c.colorMode = &on
	}
}

// WithWidth sets the terminal width for side-by-side layout and
// LayoutPreferSideBySide decisions. Default is 0, which auto-detects
// from the terminal. If detection fails (non-TTY), panels are not
// truncated. Ignored in LayoutInline mode.
func WithWidth(cols int) Option {
	return func(c *config) {
		if cols < 0 {
			cols = 0
		}
		c.width = cols
	}
}
