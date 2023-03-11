package dejavu

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func mysql(t *testing.T) (*sql.DB, PlaceholderSyntax) {
	t.Helper()

	result, err := sql.Open(
		"mysql",
		"root:root@(localhost)/deja_vu?multiStatements=true&parseTime=true",
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = result.Close()
	})

	return result, PlaceholderQuestionMark
}

func postgresql(t *testing.T) (*sql.DB, PlaceholderSyntax) {
	t.Helper()

	result, err := sql.Open("pgx", "postgresql://postgres:postgres@localhost:5432/postgres")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = result.Close()
	})

	return result, PlaceholderIndexed
}

func sqlite(t *testing.T) (*sql.DB, PlaceholderSyntax) {
	t.Helper()

	result, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = result.Close()
	})

	return result, PlaceholderQuestionMark
}

func TestDejaVu_Upgrade(t *testing.T) {
	type test struct {
		name  string
		setup func(t *testing.T) (*sql.DB, PlaceholderSyntax)
	}

	tests := []test{
		{
			name:  "sqlite",
			setup: sqlite,
		},
	}

	if !testing.Short() {
		tests = append(tests,
			test{name: "mysql", setup: mysql},
			test{name: "postgresql", setup: postgresql},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, syntax := tt.setup(t)
			dv := newTestConfig(t, db, syntax).Build()
			ctx := context.Background()

			err := dv.Upgrade(ctx)
			require.NoError(t, err)

			assert.True(t, dv.db.Exist(ctx, "deja_vu_history"))
			assert.True(t, dv.db.Exist(ctx, "deja_vu_lock"))

			count, err := dv.db.Count(ctx, "deja_vu_history")
			require.NoError(t, err)
			assert.Equal(t, 2, count)

			count, err = dv.db.Count(ctx, "deja_vu_lock")
			require.NoError(t, err)
			assert.Equal(t, 0, count)

			count, err = dv.db.Count(ctx, "country")
			require.NoError(t, err)
			assert.Equal(t, 249, count)
		})
	}
}
