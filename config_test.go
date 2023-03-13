package dejavu

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(t *testing.T, db *sql.DB, placeholders Placeholders) *Config {
	t.Helper()

	clock := newTestClock()
	logger := newTestLogger(t)

	return NewConfig(
		NewDatabase(clock, logger, NewRepository(db, logger, placeholders), DefaultStatements{}),
		newTestMigrations(t),
	).
		WithClock(clock).
		WithLogger(logger)
}

func TestNewConfig(t *testing.T) {
	logger := newTestLogger(t)
	db := NewDatabase(newTestClock(), logger, NewRepository(nil, logger, PlaceholdersQuestionMark()), DefaultStatements{})
	migs := newTestMigrations(t)
	cfg := NewConfig(db, migs)

	require.NotNil(t, cfg)
	assert.Nil(t, cfg.clock)
	assert.Equal(t, db, cfg.db)
	assert.Nil(t, cfg.logger)
	assert.Equal(t, migs, cfg.migs)
	assert.Equal(t, TickInterval, cfg.tick)
	assert.Equal(t, Timeout, cfg.timeout)
}

func TestConfig_Build(t *testing.T) {
	logger := newTestLogger(t)
	db := NewDatabase(newTestClock(), logger, NewRepository(nil, logger, PlaceholdersQuestionMark()), DefaultStatements{})
	migs := newTestMigrations(t)
	dv := NewConfig(db, migs).Build()

	require.NotNil(t, dv)
	assert.NotNil(t, dv.clock)
	assert.IsType(t, UtcClock{}, dv.clock)
	assert.Equal(t, db, dv.db)
	assert.NotNil(t, dv.logger)
	assert.IsType(t, LogLogger{}, dv.logger)
	assert.Equal(t, migs, dv.migs)
	assert.Equal(t, TickInterval, dv.tick)
	assert.Equal(t, Timeout, dv.timeout)
}

func TestConfig_WithClock(t *testing.T) {
	logger := newTestLogger(t)
	testClock := newTestClock()
	cfg := NewConfig(
		NewDatabase(newTestClock(), logger, NewRepository(nil, logger, PlaceholdersQuestionMark()), DefaultStatements{}),
		newTestMigrations(t),
	).WithClock(testClock)

	assert.Equal(t, testClock, cfg.clock)
}

func TestConfig_WithLogger(t *testing.T) {
	logger := newTestLogger(t)
	cfg := NewConfig(
		NewDatabase(newTestClock(), logger, NewRepository(nil, logger, PlaceholdersQuestionMark()), DefaultStatements{}),
		newTestMigrations(t),
	).WithLogger(logger)

	assert.Equal(t, logger, cfg.logger)
}

func TestConfig_WithTick(t *testing.T) {
	logger := newTestLogger(t)
	tick := 42 * time.Minute
	cfg := NewConfig(
		NewDatabase(newTestClock(), logger, NewRepository(nil, logger, PlaceholdersQuestionMark()), DefaultStatements{}),
		newTestMigrations(t),
	).WithTick(tick)

	assert.Equal(t, tick, cfg.tick)
}

func TestConfig_WithTimeout(t *testing.T) {
	logger := newTestLogger(t)
	timeout := 42 * time.Minute
	cfg := NewConfig(
		NewDatabase(newTestClock(), logger, NewRepository(nil, logger, PlaceholdersQuestionMark()), DefaultStatements{}),
		newTestMigrations(t),
	).WithTimeout(timeout)

	assert.Equal(t, timeout, cfg.timeout)
}

func TestConfig_String(t *testing.T) {
	assert.Equal(
		t,
		"Config: clock=Fixed clock at 2023-03-10 22:04:27 +0000 UTC, db=Database: clock=Fixed clock at 2023-03-10 22:04:27 +0000 UTC, logger=test logger, repo=SQL db with Question Mark args with ?, stmts=Default SQL statements, migs=&{testdata db}, tick=5s, timeout=5m0s", //nolint:lll
		newTestConfig(t, nil, PlaceholdersQuestionMark()).String(),
	)
}
