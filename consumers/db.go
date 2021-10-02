package consumers

import (
	"context"
	"database/sql"
	"golang.org/x/sync/errgroup"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// InsertDB creates a schema callback that will generate batches of insert statements and stream them to db.
func InsertDB(db DB, batchSize int) RawConsumer {
	return func(ctx context.Context, wg *errgroup.Group) func(t string, c []string, rows chan []string) {
		return Inserts(wg, func(statement string) error {
			_, err := db.ExecContext(ctx, statement)
			return err
		}, batchSize)
	}
}
