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

func Inserts(wg *sync.WaitGroup) Consumer {
	return func(t string, c []string, rows chan []string) {
		wg.Add(1)
		go func() {
			counter := 0

			f, err := os.Create(filepath.Join("out", t+".sql"))
			if err != nil {
				panic(err)
			}
			defer panicOnErr(f.Close)
			defer wg.Done()

			fmt.Fprintf(f, "TRUNCATE %s;\n", t)
			row := <-rows
			if len(row) > 0 {
				fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(c, ", "), strings.Join(row, ", "))
				counter += 1

				for row := range rows {
					fmt.Fprintf(f, ",\n(%s)", strings.Join(row, ", "))

					counter += 1
				}
				fmt.Fprintf(f, ";")

				fmt.Printf("%s x %d\n", t, counter)
			}
		}()
	}
}
