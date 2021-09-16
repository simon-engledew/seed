package distribution

import (
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"sync"
)

type Distribution func() bool

func Locked(fn Distribution) Distribution {
	var m sync.Mutex
	return func() bool {
		m.Lock()
		defer m.Unlock()
		return fn()
	}
}

func Fixed(size uint) Distribution {
	count := uint(0)
	return Locked(func() bool {
		count += 1
		return count <= size
	})
}

func Ratio(ratio float64) Distribution {
	var done bool
	return Locked(func() bool {
		if done {
			return false
		}
		done = true
		return rand.Float64() <= ratio
	})
}

func Range(min, max uint) Distribution {
	count := uint(0)
	size := uint(gofakeit.Number(int(min), int(max)))
	return Locked(func() bool {
		count += 1
		return count <= size
	})
}
