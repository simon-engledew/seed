package distribution

type Distribution func() uint

func Fixed(size uint) Distribution {
	return func() uint {
		return size
	}
}
