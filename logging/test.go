package logging

import "testing"

// TestLogger logs everything through a testing.T.
type TestLogger struct {
	t *testing.T
}

// Test logger.
func Test(t *testing.T) *TestLogger {
	t.Helper()

	return &TestLogger{t}
}

// Debugf calls t.Logf.
func (l *TestLogger) Debugf(format string, args ...any) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

// Infof calls t.Logf.
func (l *TestLogger) Infof(format string, args ...any) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

// Warnf calls t.Logf.
func (l *TestLogger) Warnf(format string, args ...any) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

// Errorf calls t.Logf.
func (l *TestLogger) Errorf(format string, args ...any) {
	l.t.Helper()
	l.t.Logf(format, args...)
}
