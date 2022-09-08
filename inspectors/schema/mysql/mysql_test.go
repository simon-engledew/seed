package mysql_test

import (
	"github.com/simon-engledew/seed/inspectors/schema/mysql"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestExample(t *testing.T) {
	tables, err := mysql.InspectMySQLSchema(strings.NewReader(`
	CREATE TABLE owners (
		id BIGINT UNSIGNED,
		name VARCHAR(255)
	);

    CREATE TABLE cats (
		id BIGINT UNSIGNED,
		owner_id BIGINT,
		name VARCHAR(255)
    );
	`))
	require.NoError(t, err)
	require.Equal(t, tables, []byte(`{"cats":[{"name":"id","data_type":"bigint","is_primary":false,"is_unsigned":true,"length":20,"column_type":"bigint(20)"},{"name":"owner_id","data_type":"bigint","is_primary":false,"is_unsigned":false,"length":20,"column_type":"bigint(20)"},{"name":"name","data_type":"varchar","is_primary":false,"is_unsigned":false,"length":255,"column_type":"varchar(255)"}],"owners":[{"name":"id","data_type":"bigint","is_primary":false,"is_unsigned":true,"length":20,"column_type":"bigint(20)"},{"name":"name","data_type":"varchar","is_primary":false,"is_unsigned":false,"length":255,"column_type":"varchar(255)"}]}`))
}
