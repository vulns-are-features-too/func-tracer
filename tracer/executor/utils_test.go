package executor_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vulns-are-features-too/func-tracer/tracer/executor"
)

func TestNextCapacity(t *testing.T) {
	t.Parallel()

	const iMax = math.MaxInt

	testcases := []struct{ in, out int }{
		{math.MinInt, 1},
		{-1, 1},
		{0, 1},
		{1, 2},
		{2, 4},
		{3, 4},
		{4, 8},
		{5, 8},
		{7, 8},
		{8, 16},
		{9, 16},
		{15, 16},
		{16, 32},
		{17, 32},
		{iMax, iMax},
		{iMax - 1, iMax},
		{(iMax >> 1) + 1, iMax},      // 0010... => 0100...
		{iMax >> 1, (iMax >> 1) + 1}, // 0011... => 0100...
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%d=>%d", tc.in, tc.out), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.out, executor.NextCapacity(tc.in))
		})
	}

	t.Run("ALL i < NextCapacity(i) <= (i<<1) WHERE 0 < i < max", func(t *testing.T) {
		t.Parallel()

		check := func(i int) {
			res := executor.NextCapacity(i)

			// i < InitialCapacity(i)
			assert.Greater(t, res, i)

			if (i << 1) > 0 {
				// InitialCapacity(i) <= (i<<1)
				assert.LessOrEqual(t, res, i<<1)
			}
		}

		// actual ALL takes too long,
		// so just checking around powers of 2
		for i := 1; i > 0; i <<= 1 {
			check(i - 1)
			check(i)
			check(i + 1)
		}
	})
}
