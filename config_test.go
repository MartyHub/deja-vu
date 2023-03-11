package dejavu

import (
	"database/sql"
	"testing"
)

func newTestConfig(t *testing.T, db *sql.DB, syntax PlaceholderSyntax) *Config {
	t.Helper()

	clock := newTestClock()
	logger := newTestLogger(t)
	repo := NewRepository(db, logger, syntax)

	return NewConfig(
		NewDatabase(clock, logger, repo, DefaultStatements{}),
		newTestMigrations(t),
	).
		WithClock(clock).
		WithLogger(logger)
}
