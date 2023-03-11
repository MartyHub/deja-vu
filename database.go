package dejavu

import (
	"context"
	"fmt"
)

type Database interface {
	fmt.Stringer

	Ping(ctx context.Context) error

	Count(ctx context.Context, table string) (int, error)
	Exist(ctx context.Context, table string) bool

	InitLockTable(ctx context.Context) error
	InitHistoryTable(ctx context.Context) error

	Lock(ctx context.Context, lck Lock) bool

	History(ctx context.Context) ([]Migration, error)
	Migrate(ctx context.Context, name, content string) error

	Unlock(ctx context.Context, lck Lock) error
}

func NewDatabase(clock Clock, logger Logger, repo Repository, stmts Statements) DefaultDatabase {
	return DefaultDatabase{
		clock:  clock,
		logger: logger,
		repo:   repo,
		stmts:  stmts,
	}
}

type DefaultDatabase struct {
	clock  Clock
	logger Logger
	repo   Repository
	stmts  Statements
}

func (d DefaultDatabase) Ping(ctx context.Context) error {
	if err := d.repo.Ping(ctx); err != nil {
		return newError(err, "failed to ping database")
	}

	return nil
}

func (d DefaultDatabase) Count(ctx context.Context, table string) (int, error) {
	row := d.repo.QueryRow(ctx, d.stmts.CountFromTable(table))

	var result int
	if err := row.Scan(&result); err != nil {
		return 0, newError(err, "failed to count table %s", table)
	}

	return result, nil
}

func (d DefaultDatabase) Exist(ctx context.Context, table string) bool {
	_, err := d.Count(ctx, table)

	return err == nil
}

func (d DefaultDatabase) InitLockTable(ctx context.Context) error {
	if !d.Exist(ctx, LockTableName) {
		d.logger.Log("Creating lock table...")

		if err := d.repo.EnsureTransaction(ctx, nil, func(ctx context.Context, repo Repository) error {
			return repo.Exec(ctx, d.stmts.CreateLockTable())
		}); err != nil {
			return newError(err, "failed to create lock table")
		}

		d.logger.Log("Lock table successfully created")
	}

	return nil
}

func (d DefaultDatabase) InitHistoryTable(ctx context.Context) error {
	if !d.Exist(ctx, HistoryTableName) {
		d.logger.Log("Creating history table...")

		err := d.repo.EnsureTransaction(ctx, nil, func(ctx context.Context, repo Repository) error {
			return repo.Exec(ctx, d.stmts.CreateHistoryTable())
		})
		if err != nil {
			return newError(err, "failed to create history table")
		}

		d.logger.Log("History table successfully created")
	}

	return nil
}

func (d DefaultDatabase) Lock(ctx context.Context, lck Lock) bool {
	d.logger.Log("Acquiring lock...")

	err := d.repo.EnsureTransaction(ctx, nil, func(ctx context.Context, repo Repository) error {
		return repo.Exec(ctx, d.stmts.Lock(lck))
	})
	if err == nil {
		d.logger.Log("Lock successfully acquired")

		return true
	}

	return false
}

func (d DefaultDatabase) History(ctx context.Context) ([]Migration, error) {
	d.logger.Log("Finding existing migrations...")

	rows, err := d.repo.Query(ctx, d.stmts.History())
	if err != nil {
		return nil, newError(err, "failed to query database history")
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	migs := make([]Migration, 0)

	for rows.Next() {
		var mig Migration

		if err = rows.Scan(&mig.Name, &mig.Start, &mig.DurationMs, &mig.Checksum); err != nil {
			return migs, newError(err, "failed to scan database history")
		}

		migs = append(migs, mig)
	}

	d.logger.Log(fmt.Sprintf("Found %d existing migration(s)", len(migs)))

	return migs, nil
}

func (d DefaultDatabase) Migrate(ctx context.Context, name, content string) error {
	start := d.clock.Now()

	if err := d.repo.EnsureTransaction(ctx, nil, func(ctx context.Context, repo Repository) error {
		return repo.Exec(ctx, NewStatement(content))
	}); err != nil {
		return newError(err, "migration %s failed", name)
	}

	duration := d.clock.Now().Sub(start)

	return d.repo.EnsureTransaction(ctx, nil, func(ctx context.Context, repo Repository) error {
		mig := Migration{
			Name:       name,
			Start:      start,
			DurationMs: duration.Milliseconds(),
			Checksum:   checksum(content),
		}

		if err := repo.Exec(ctx, d.stmts.Log(mig)); err != nil {
			return newError(err, "failed to save migration %s", mig.Name)
		}

		return nil
	})
}

func (d DefaultDatabase) Unlock(ctx context.Context, lck Lock) error {
	d.logger.Log("Freeing lock...")

	err := d.repo.EnsureTransaction(ctx, nil,
		func(ctx context.Context, repo Repository) error {
			return repo.Exec(ctx, d.stmts.Unlock(lck))
		},
	)
	if err != nil {
		return newError(err, "failed to free lock")
	}

	d.logger.Log("Lock successfully freed")

	return nil
}

func (d DefaultDatabase) String() string {
	return fmt.Sprintf("Database: clock=%v, logger=%v, repo=%v, stmts=%v",
		d.clock,
		d.logger,
		d.repo,
		d.stmts,
	)
}
