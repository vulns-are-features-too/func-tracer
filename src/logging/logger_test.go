//go:build !race

package logging_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vulns-are-features-too/func-tracer/src/logging"
)

func TestImplLogger(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*logging.Logger)(nil), new(logging.CliLogger))
	assert.Implements(t, (*logging.Logger)(nil), new(logging.TestLogger))
}
