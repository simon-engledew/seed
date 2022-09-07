package consumers

import (
	"context"
	"github.com/simon-engledew/seed/generators"
	"golang.org/x/sync/errgroup"
)

// Consumer returns a callback that takes generated rows and turns them into a format the database can import.
type Consumer func(context.Context, *errgroup.Group) func(t string, c []string, rows chan []*generators.Value)
type rawConsumer func(context.Context, *errgroup.Group) func(t string, c []string, rows chan []string)
