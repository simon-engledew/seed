package consumers

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"sync"
)

// InsertWriter creates a schema callback that will generate batches of insert statements and stream them to w.
func InsertWriter(w io.Writer, batchSize int) rawConsumer {
	var mutex sync.Mutex
	return func(ctx context.Context, wg *errgroup.Group) func(t string, c []string, rows chan []string) {
		return Inserts(wg, func(statement string) error {
			mutex.Lock()
			_, err := fmt.Fprint(w, statement+";\n")
			mutex.Unlock()
			return err
		}, batchSize)
	}
}
