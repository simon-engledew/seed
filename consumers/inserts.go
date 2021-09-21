package consumers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func panicOnErr(fn func() error) {
	if err := fn(); err != nil {
		panic(err)
	}
}

func Inserts(dir string, batchSize int) Consumer {
	return func(wg *sync.WaitGroup) func(t string, c []string, rows chan []string) {
		return func(t string, c []string, rows chan []string) {
			wg.Add(1)
			go func() {
				counter := 0

				f, err := os.Create(filepath.Join(dir, t+".sql"))
				if err != nil {
					panic(err)
				}
				defer panicOnErr(f.Close)
				defer wg.Done()

				fmt.Fprintf(f, "TRUNCATE %s;\n", t)
				for row := <-rows; row != nil; row = <-rows {
					fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(c, ", "), strings.Join(row, ", "))
					counter += 1

					for i := 0; i < batchSize-1; i += 1 {
						row = <-rows
						if row == nil {
							break
						}

						fmt.Fprintf(f, ",\n(%s)", strings.Join(row, ", "))

						counter += 1
					}
					fmt.Fprintf(f, ";\n")
				}

				fmt.Printf("%s x %d\n", t, counter)
			}()
		}
	}
}
