package consumers

import (
	"bytes"
	"fmt"
	"golang.org/x/sync/errgroup"
	"strings"
)

func Inserts(wg *errgroup.Group, fn func(stmt string) error, batchSize int) func(t string, c []string, rows chan []string) {
	return func(t string, c []string, rows chan []string) {
		wg.Go(func() (err error) {
			var buf bytes.Buffer

			w := &buf

			err = fn(`SET autocommit = 0`)
			if err != nil {
				return
			}

			err = fn(`TRUNCATE ` + t)
			if err != nil {
				return
			}

			for row := <-rows; row != nil; row = <-rows {
				_, err = fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(c, ", "), strings.Join(row, ", "))
				if err != nil {
					return
				}

				i := 1
				for ; i < batchSize; i += 1 {
					row = <-rows
					if row == nil {
						break
					}

					_, err = fmt.Fprintf(w, ",\n(%s)", strings.Join(row, ", "))
					if err != nil {
						return
					}
				}

				_, err = fmt.Fprintf(w, ";\n")
				if err != nil {
					return
				}

				err = fn(buf.String())
				if err != nil {
					return
				}

				buf.Reset()
			}

			return fn(`COMMIT`)
		})
	}
}
