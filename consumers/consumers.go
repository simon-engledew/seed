package consumers

import (
	"golang.org/x/sync/errgroup"
)

type Consumer func(*errgroup.Group) func(t string, c []string, rows chan []string)
