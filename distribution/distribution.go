package distribution

import (
	"math/rand"
)

type Distribution func() bool

func Fixed(size uint) Distribution {
	count := uint(0)
	return func() bool {
		count += 1
		return count <= size
	}
}

func Ratio(ratio float64) Distribution {
	var done bool
	return func() bool {
		if done {
			return false
		}
		done = true
		return rand.Float64() <= ratio
	}
}
