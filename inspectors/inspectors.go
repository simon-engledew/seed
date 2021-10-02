package inspectors

import "github.com/simon-engledew/seed/generators"

// Inspector returns a function that can provide metadata about the schema we are building test data for.
type Inspector func() (map[string]map[string]Column, error)

type Column interface {
	Generator() generators.ValueGenerator
	Type() string
}
