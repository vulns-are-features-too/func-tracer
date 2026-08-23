package cache_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/src/tracer/cache"
)

func TestGetOrExecOnlyExecsOnce(t *testing.T) {
	t.Parallel()

	cache := cache.New[any]()
	count := 0

	var wg sync.WaitGroup

	for range 32 {
		wg.Go(func() {
			_, err := cache.GetOrExec("k", func() (any, error) {
				count++

				return struct{}{}, nil
			})

			require.NoError(t, err)
		})
	}

	wg.Wait()

	assert.Equal(t, 1, count)
}
