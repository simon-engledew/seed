package consumers

import (
	"bytes"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"strings"
	"sync"
)

func Inserts(w io.Writer, batchSize int) Consumer {
	var mutex sync.Mutex
	return func(wg *errgroup.Group) func(t string, c []string, rows chan []string) {
		return func(t string, c []string, rows chan []string) {
			wg.Go(func() (err error) {
				var buf bytes.Buffer

				mutex.Lock()
				_, err = fmt.Fprintf(w, "TRUNCATE %s;\n", t)
				mutex.Unlock()
				if err != nil {
					return
				}

				for row := <-rows; row != nil; row = <-rows {
					_, err = fmt.Fprintf(&buf, "INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(c, ", "), strings.Join(row, ", "))
					if err != nil {
						return
					}

					for i := 1; i < batchSize; i += 1 {
						row = <-rows
						if row == nil {
							break
						}

						_, err = fmt.Fprintf(&buf, ",\n(%s)", strings.Join(row, ", "))
						if err != nil {
							return
						}
					}

					_, err = fmt.Fprintf(&buf, ";\n")
					if err != nil {
						return
					}

					mutex.Lock()
					_, err = io.Copy(w, &buf)
					mutex.Unlock()
					if err != nil {
						return
					}
				}

				return
			})
		}
	}
}
