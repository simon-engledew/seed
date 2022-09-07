package mysql_schema

import (
	"fmt"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/mysql"
	"github.com/pingcap/tidb/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed/inspectors"
	"io"
)

func InspectMySQLSchema(r io.Reader) inspectors.Inspector {
	p := parser.New()

	return func(fn func(tableName, columnName string, column inspectors.ColumnInfo)) error {
		dump, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		statements, _, err := p.Parse(string(dump), "", "")
		if err != nil {
			return fmt.Errorf("failed to parse sqldump: %w", err)
		}

		for _, statement := range statements {
			if create, ok := statement.(*ast.CreateTableStmt); ok {
				tableName := create.Table.Name.String()

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

					fn(tableName, columnName, inspectors.MySQLColumn{
						Name:       columnName,
						IsPrimary:  isPrimary,
						IsUnsigned: mysql.HasUnsignedFlag(col.Tp.GetFlag()),
						ColumnType: col.Tp.CompactStr(),
						DataType:   types.TypeToStr(col.Tp.GetType(), col.Tp.GetCharset()),
						Length:     length,
					})
				}
			}
		}

		return nil
	}
}
