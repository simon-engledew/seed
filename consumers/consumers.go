package consumers

import (
	"golang.org/x/sync/errgroup"
)

// Consumer returns a callback that takes generated rows and turns them into a format the database can import.
type Consumer func(*errgroup.Group) func(t string, c []string, rows chan []string)
