package logging

import (
	"fmt"
	"log"
)

// CliLogger is for logging when run as a CLI.
type CliLogger struct {
	lvl    Level
	logger *log.Logger
}

// New CLI logger.
func New(lvl Level) *CliLogger {
	return &CliLogger{lvl, log.Default()}
}

// Debugf logs at debug level.
func (l *CliLogger) Debugf(format string, args ...any) {
	l.logf(LevelDebug, "DEBUG", format, args...)
}

// Infof logs at info level.
func (l *CliLogger) Infof(format string, args ...any) {
	l.logf(LevelInfo, "INFO", format, args...)
}

// Warnf logs at warn level.
func (l *CliLogger) Warnf(format string, args ...any) {
	l.logf(LevelWarn, "WARN", format, args...)
}

// Errorf logs at error level.
func (l *CliLogger) Errorf(format string, args ...any) {
	l.logf(LevelError, "ERROR", format, args...)
}

func (l *CliLogger) logf(lvl Level, lvlTxt string, format string, args ...any) {
	if lvl < l.lvl {
		return
	}

	s := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s\n", lvlTxt, s)
}
