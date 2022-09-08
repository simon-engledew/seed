package mysql_test

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/simon-engledew/seed/inspectors/db/mysql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExample(t *testing.T) {
	db, err := sql.Open("mysql", "root@/information_schema")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	_, err = db.Exec(`
	CREATE DATABASE IF NOT EXISTS seed CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
 	`)
	_, err = db.Exec(`
	USE seed
 	`)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS owners (
		id BIGINT UNSIGNED,
		name VARCHAR(255)
	)
	`)
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS cats (
		id BIGINT UNSIGNED,
		owner_id BIGINT,
		name VARCHAR(255)
    )
	`)
	require.NoError(t, err)
	tables, err := mysql.InspectMySQLConnection(db)
	require.NoError(t, err)
	require.Equal(t, tables, []byte(`{"cats": [{"name": "id", "length": null, "data_type": "bigint", "is_primary": false, "column_type": "bigint unsigned", "is_unsigned": true}, {"name": "name", "length": 255, "data_type": "varchar", "is_primary": false, "column_type": "varchar(255)", "is_unsigned": false}, {"name": "owner_id", "length": null, "data_type": "bigint", "is_primary": false, "column_type": "bigint", "is_unsigned": false}], "owners": [{"name": "id", "length": null, "data_type": "bigint", "is_primary": false, "column_type": "bigint unsigned", "is_unsigned": true}, {"name": "name", "length": 255, "data_type": "varchar", "is_primary": false, "column_type": "varchar(255)", "is_unsigned": false}]}`))
}
