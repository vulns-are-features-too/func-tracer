// Package logging provides logging swappable based on context
package logging

// Logger writes logs to underlying sink depending on implementation.
type Logger interface {
	// Debugf logs at debug level
	Debugf(format string, args ...any)
	// Infof logs at info level
	Infof(format string, args ...any)
	// Warnf logs at warn level
	Warnf(format string, args ...any)
	// Errorf logs at error level
	Errorf(format string, args ...any)
}
