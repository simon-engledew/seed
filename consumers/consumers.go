package consumers

import (
	"context"
	"golang.org/x/sync/errgroup"
)

type ValueGenerator interface {
	Value(ctx context.Context) Value
}

type Value interface {
	String() string
	Escape() bool
}

// Consumer returns a callback that takes generated rows and turns them into a format the database can import.
type Consumer func(context.Context, *errgroup.Group) func(t string, c []string, rows chan []Value)
type rawConsumer func(context.Context, *errgroup.Group) func(t string, c []string, rows chan []string)
