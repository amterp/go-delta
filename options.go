package godelta

// Option configures the behavior of DiffWith.
type Option func(*config)

type config struct {
	contextLines int
	sideBySide   bool
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

// WithSideBySide enables or disables side-by-side layout.
// Default is false (inline).
func WithSideBySide(on bool) Option {
	return func(c *config) {
		c.sideBySide = on
	}
}

// WithColor forces color on or off, overriding auto-detection.
func WithColor(on bool) Option {
	return func(c *config) {
		c.colorMode = &on
	}
}

// WithWidth sets the terminal width used for side-by-side layout.
// Default is 0, which auto-detects from the terminal (fallback 80).
// Ignored in inline mode.
func WithWidth(cols int) Option {
	return func(c *config) {
		if cols < 0 {
			cols = 0
		}
		c.width = cols
	}
}
