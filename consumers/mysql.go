package consumers

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/simon-engledew/seed/generators"
	"golang.org/x/sync/errgroup"
	"io"
	"strings"
)

func MySQLConsumer(base RawConsumer) Consumer {
	return func(ctx context.Context, wg *errgroup.Group) func(t string, c []string, rows chan []*generators.Value) {
		fn := base(ctx, wg)
		return func(tableName string, cols []string, rows chan []*generators.Value) {
			quotedCols := make([]string, len(cols))
			for n, col := range cols {
				quotedCols[n] = QuoteIdentifier(col)
			}
			quotedRows := make(chan []string)
			wg.Go(func() error {
				for row := <-rows; row != nil; row = <-rows {
					quotedRow := make([]string, len(row))
					for n := range row {
						quotedRow[n] = Quote(row[n])
					}
					quotedRows <- quotedRow
				}

				return nil
			})
			fn(QuoteIdentifier(tableName), quotedCols, quotedRows)
		}
	}
}

func MySQLInsertWriter(w io.Writer, batchSize int) Consumer {
	return MySQLConsumer(InsertWriter(w, batchSize))
}

func MySQLInsertDB(db *sql.DB, batchSize int) Consumer {
	return MySQLConsumer(InsertDB(db, batchSize))
}

func Quote(val *generators.Value) string {
	if val.Quote {
		return quote(val.Value)
	}
	return val.Value
}

func quote(str string) string {
	runes := []rune(str)
	buffer := bytes.NewBufferString("")
	buffer.WriteRune('\'')
	for i, runeLength := 0, len(runes); i < runeLength; i++ {
		switch runes[i] {
		case '\\', '\'':
			buffer.WriteRune('\\')
			buffer.WriteRune(runes[i])
		case 0:
			buffer.WriteRune('\\')
			buffer.WriteRune('0')
		case '\032':
			buffer.WriteRune('\\')
			buffer.WriteRune('Z')
		default:
			buffer.WriteRune(runes[i])
		}
	}
	buffer.WriteRune('\'')

	return buffer.String()
}

func QuoteIdentifier(v string) string {
	return "`" + strings.ReplaceAll(v, "`", "``") + "`"
}
