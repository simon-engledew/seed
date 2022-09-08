package mysql

import (
	"fmt"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/mysql"
	"github.com/pingcap/tidb/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/generators"
	"io"
)

type MySQLColumn struct {
	Name       string
	DataType   string
	IsPrimary  bool
	IsUnsigned bool
	Length     int
	Type       string
}

func toColumn(c MySQLColumn) *seed.Column {
	return &seed.Column{
		Name:      c.Name,
		Type:      c.Type,
		Generator: toGenerator(c),
	}
}

func toGenerator(c MySQLColumn) generators.ValueGenerator {
	if c.IsPrimary {
		return generators.Counter()
	}
	gen := generators.Column(c.DataType, c.IsUnsigned, c.Length)
	if gen == nil {
		return generators.Identity(c.Type, true)
	}
	return gen
}

func InspectMySQLSchema(r io.Reader) (map[string][]*seed.Column, error) {
	p := parser.New()

	dump, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	statements, _, err := p.Parse(string(dump), "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse sqldump: %w", err)
	}

	schema := make(map[string][]*seed.Column)

	for _, statement := range statements {
		if create, ok := statement.(*ast.CreateTableStmt); ok {
			tableName := create.Table.Name.String()

			schema[tableName] = make([]*seed.Column, 0, len(create.Cols))

			primaryKey := make(map[string]struct{})

			for _, constraint := range create.Constraints {
				if constraint.Tp == ast.ConstraintPrimaryKey {
					for _, key := range constraint.Keys {
						columnName := key.Column.String()

						primaryKey[columnName] = struct{}{}
					}
				}
			}

		outer:
			for _, col := range create.Cols {
				columnName := col.Name.String()

				for _, option := range col.Options {
					if option.Tp == ast.ColumnOptionGenerated {
						continue outer
					}
					if option.Tp == ast.ColumnOptionPrimaryKey {
						primaryKey[columnName] = struct{}{}
					}
				}

				_, isPrimary := primaryKey[columnName]

				length := col.Tp.GetFlen()
				if length == types.UnspecifiedLength {
					length, _ = mysql.GetDefaultFieldLengthAndDecimal(col.Tp.GetType())
				}

				schema[tableName] = append(schema[tableName], toColumn(MySQLColumn{
					Name:       columnName,
					IsPrimary:  isPrimary,
					IsUnsigned: mysql.HasUnsignedFlag(col.Tp.GetFlag()),
					Type:       col.Tp.CompactStr(),
					DataType:   types.TypeToStr(col.Tp.GetType(), col.Tp.GetCharset()),
					Length:     length,
				}))
			}
		}
	}

	return schema, nil
}
