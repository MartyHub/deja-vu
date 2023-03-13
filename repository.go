package dejavu

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	fmt.Stringer

	EnsureTransaction(
		ctx context.Context,
		opts *sql.TxOptions,
		f func(ctx context.Context, repo Repository) error,
	) error

	Ping(ctx context.Context) error
	Exec(ctx context.Context, stmt *Statement) error
	Query(ctx context.Context, stmt *Statement) (*sql.Rows, error)
	QueryRow(ctx context.Context, stmt *Statement) *sql.Row
}

func NewRepository(db *sql.DB, logger Logger, placeholders Placeholders) DBRepository {
	return DBRepository{
		db:           db,
		logger:       logger,
		placeholders: placeholders,
	}
}

type DBRepository struct {
	db           *sql.DB
	logger       Logger
	placeholders Placeholders
}

//nolint:nakedret
func (repo DBRepository) EnsureTransaction(
	ctx context.Context,
	opts *sql.TxOptions,
	f func(ctx context.Context, repo Repository) error,
) (err error) {
	repo.logger.Log("Starting transaction...")

	tx, err := repo.db.BeginTx(ctx, opts)
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			repo.logger.Log("Rollbacking transaction...")

			_ = tx.Rollback()

			panic(p)
		} else if err != nil {
			repo.logger.Log("Rollbacking transaction...")
			_ = tx.Rollback()
		} else {
			repo.logger.Log("Committing transaction...")
			err = tx.Commit()
		}
	}()

	err = f(ctx, txRepository{
		DBRepository: repo,
		tx:           tx,
		opts:         opts,
	})

	return
}

func (repo DBRepository) Ping(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

func (repo DBRepository) Exec(ctx context.Context, stmt *Statement) error {
	query, args := stmt.WithPlaceholders(repo.placeholders)
	logStatement(repo.logger, query, args)
	_, err := repo.db.ExecContext(ctx, query, args...)

	return err
}

func (repo DBRepository) Query(ctx context.Context, stmt *Statement) (*sql.Rows, error) {
	query, args := stmt.WithPlaceholders(repo.placeholders)
	logStatement(repo.logger, query, args)

	return repo.db.QueryContext(ctx, query, args...)
}

func (repo DBRepository) QueryRow(ctx context.Context, stmt *Statement) *sql.Row {
	query, args := stmt.WithPlaceholders(repo.placeholders)
	logStatement(repo.logger, query, args)

	return repo.db.QueryRowContext(ctx, query, args...)
}

func (repo DBRepository) String() string {
	return fmt.Sprintf("SQL db with %v", repo.placeholders)
}

type txRepository struct {
	DBRepository
	tx   *sql.Tx
	opts *sql.TxOptions
}

func (repo txRepository) EnsureTransaction(
	ctx context.Context,
	opts *sql.TxOptions,
	f func(ctx context.Context, repo Repository) error,
) error {
	if opts == nil || repo.opts != nil && *repo.opts == *opts {
		repo.logger.Log("Already in a transaction")

		return f(ctx, repo)
	}

	// Create new transaction
	return repo.DBRepository.EnsureTransaction(ctx, opts, f)
}

func (repo txRepository) Ping(ctx context.Context) error {
	return repo.DBRepository.Ping(ctx)
}

func (repo txRepository) Exec(ctx context.Context, stmt *Statement) error {
	query, args := stmt.WithPlaceholders(repo.placeholders)
	logStatement(repo.logger, query, args)
	_, err := repo.tx.ExecContext(ctx, query, args...)

	return err
}

func (repo txRepository) Query(ctx context.Context, stmt *Statement) (*sql.Rows, error) {
	query, args := stmt.WithPlaceholders(repo.placeholders)
	logStatement(repo.logger, query, args)

	return repo.tx.QueryContext(ctx, query, args...)
}

func (repo txRepository) QueryRow(ctx context.Context, stmt *Statement) *sql.Row {
	query, args := stmt.WithPlaceholders(repo.placeholders)
	logStatement(repo.logger, query, args)

	return repo.tx.QueryRowContext(ctx, query, args...)
}

func (repo txRepository) String() string {
	return fmt.Sprintf("SQL tx with %v args", repo.placeholders)
}

func logStatement(logger Logger, query string, args []any) {
	logger.Log("Statement: " + query)

	for i, arg := range args {
		logger.Log(fmt.Sprintf("Arg %d: %v", i+1, arg))
	}
}
