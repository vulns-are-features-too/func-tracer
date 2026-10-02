package executor

import (
	"math"
	"math/bits"
)

// NextCapacity returns the first power of 2 greater than required
// If required is <= 1, return 1 for safety.
func NextCapacity(required int) int {
	if required <= 0 {
		return 1
	}

	// 2nd left-most bit set => there's no greater power of 2
	//nolint:mnd
	if (required >> (bits.UintSize - 2)) != 0 {
		return math.MaxInt
	}

	return 1 << bits.Len(uint(required))
}
