package consumers

import (
	"context"
	"database/sql"
	"golang.org/x/sync/errgroup"
	"io"
	"strings"
)

func MySQLConsumer(base Consumer) Consumer {
	return func(ctx context.Context, wg *errgroup.Group) func(t string, c []string, rows chan []string) {
		fn := base(ctx, wg)
		return func(tableName string, cols []string, rows chan []string) {
			quotedCols := make([]string, len(cols))
			for n, col := range cols {
				quotedCols[n] = QuoteIdentifier(col)
			}
			fn(QuoteIdentifier(tableName), quotedCols, rows)
		}
	}
}

func MySQLInsertWriter(w io.Writer, batchSize int) Consumer {
	return MySQLConsumer(InsertWriter(w, batchSize))
}

func MySQLInsertDB(db *sql.DB, batchSize int) Consumer {
	return MySQLConsumer(InsertDB(db, batchSize))
}

func QuoteIdentifier(v string) string {
	return "`" + strings.ReplaceAll(v, "`", "``") + "`"
}
