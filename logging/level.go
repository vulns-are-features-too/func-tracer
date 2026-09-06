package logging

// Level for logging.
type Level int

const (
	// LevelDebug for debugging info.
	LevelDebug Level = 0
	// LevelInfo for informative messages.
	LevelInfo Level = 1
	// LevelWarn for warnings.
	LevelWarn Level = 2
	// LevelError for errors.
	LevelError Level = 4
)

// LevelFromVerbosity returns the logging level
// based on verbosity in the CLI.
func LevelFromVerbosity(verbosity int) Level {
	//nolint:mnd
	switch verbosity {
	case 0:
		return LevelError
	case 1:
		return LevelWarn
	case 2:
		return LevelInfo
	case 3:
		return LevelDebug
	default:
		if verbosity >= 3 {
			return LevelDebug
		}

		return LevelError
	}
}
