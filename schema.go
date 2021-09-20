package seed

import "context"

type Schema map[string][]*Column

func (s Schema) Generator(cb func(table string, columns []string, rows chan []string)) Generator {
	channels := make(map[string]chan []string)
	for t, columns := range s {
		channel := make(chan []string)
		channels[t] = channel

		names := make([]string, 0, len(columns))

		for _, column := range columns {
			names = append(names, column.Name)
		}

		cb(t, names, channel)
	}
	return &insertStack{
		ctx:      context.Background(),
		cb:       cb,
		channels: channels,
		schema:   s,
		stack:    make(Rows),
	}
}

type SchemaTransform func(t string, c *Column)

func (s Schema) Transform(transform SchemaTransform) {
	for t, columns := range s {
		for _, c := range columns {
			transform(t, c)
		}
	}
}

func (s Schema) Merge(other Schema) {
	for t, columns := range other {
		for c, g := range columns {
			s[t][c] = g
		}
	}
}
