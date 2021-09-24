package consumers_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/simon-engledew/seed/consumers"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"strconv"
	"testing"
)

func TestInsertsConsumerShort(t *testing.T) {
	var buf bytes.Buffer
	group, _ := errgroup.WithContext(context.Background())

	rows := make(chan []string)

	go func() {
		for i := 1; i <= 2; i++ {
			rows <- []string{strconv.Itoa(i), fmt.Sprintf("'test-%d'", i)}
		}
		close(rows)
	}()

	fn := consumers.Inserts(&buf, 5)(group)
	fn("test", []string{"id", "name"}, rows)

	require.NoError(t, group.Wait())
	require.Equal(t, "TRUNCATE test;\nINSERT INTO test (id, name) VALUES (1, 'test-1'),\n(2, 'test-2');\n", buf.String())
}

func TestInsertsConsumerLong(t *testing.T) {
	var buf bytes.Buffer
	group, _ := errgroup.WithContext(context.Background())

	rows := make(chan []string)

	go func() {
		for i := 1; i <= 2; i++ {
			rows <- []string{strconv.Itoa(i), fmt.Sprintf("'test-%d'", i)}
		}
		close(rows)
	}()

	fn := consumers.Inserts(&buf, 1)(group)
	fn("test", []string{"id", "name"}, rows)

	require.NoError(t, group.Wait())
	require.Equal(t, "TRUNCATE test;\nINSERT INTO test (id, name) VALUES (1, 'test-1');\nINSERT INTO test (id, name) VALUES (2, 'test-2');\n", buf.String())
}

func TestInsertsConsumerNone(t *testing.T) {
	var buf bytes.Buffer
	group, _ := errgroup.WithContext(context.Background())

	rows := make(chan []string)

	close(rows)

	fn := consumers.Inserts(&buf, 1)(group)
	fn("test", []string{"id", "name"}, rows)

	require.NoError(t, group.Wait())
	require.Equal(t, "TRUNCATE test;\n", buf.String())
}

func TestInsertsConsumerInterleaved(t *testing.T) {
	var buf bytes.Buffer
	group, _ := errgroup.WithContext(context.Background())

	dogs := make(chan []string, 2)
	cats := make(chan []string, 2)

	dogs <- []string{"1", "'doggo'"}
	dogs <- []string{"2", "'grizzlechops'"}

	cats <- []string{"1", "'paws'"}
	cats <- []string{"2", "'sad tony'"}

	fn := consumers.Inserts(&buf, 5)(group)
	fn("dogs", []string{"id", "name"}, dogs)
	fn("cats", []string{"id", "name"}, cats)

	close(dogs)
	close(cats)

	require.NoError(t, group.Wait())
	require.Contains(t, buf.String(), "TRUNCATE dogs;\n")
	require.Contains(t, buf.String(), "TRUNCATE cats;\n")
	require.Contains(t, buf.String(), "INSERT INTO cats (id, name) VALUES (1, 'paws'),\n(2, 'sad tony');\n")
	require.Contains(t, buf.String(), "INSERT INTO dogs (id, name) VALUES (1, 'doggo'),\n(2, 'grizzlechops');\n")
}
