package dejavu

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Error struct {
	Cause   error
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func newError(cause error, format string, args ...any) error {
	return &Error{
		Cause:   cause,
		Message: fmt.Sprintf(format, args...),
	}
}

type DejaVu struct {
	Config
}

func (dv DejaVu) History(ctx context.Context) ([]Migration, error) {
	return dv.db.History(ctx)
}

//nolint:nakedret
func (dv DejaVu) Upgrade(ctx context.Context) (err error) {
	dv.logger.Log("Starting database upgrade...")

	if err = dv.init(ctx); err != nil {
		return
	}

	var lck Lock

	lck, err = dv.lock(ctx)
	if err != nil {
		return
	}

	defer func() {
		if e := dv.db.Unlock(ctx, lck); e != nil {
			if err == nil {
				err = e
			} else {
				dv.logger.Log(fmt.Sprintf("failed to free lock: %v", e))
			}
		}
	}()

	if err = dv.doUpgrade(ctx); err != nil {
		return
	}

	dv.logger.Log("Database successfully upgraded")

	return
}

func (dv DejaVu) init(ctx context.Context) error {
	if err := dv.db.Ping(ctx); err != nil {
		return err
	}

	if err := dv.db.InitLockTable(ctx); err != nil {
		return err
	}

	if err := dv.db.InitHistoryTable(ctx); err != nil {
		return err
	}

	return nil
}

func (dv DejaVu) doUpgrade(ctx context.Context) error {
	history, err := dv.db.History(ctx)
	if err != nil {
		return err
	}

	migs, err := dv.migs.List()
	if err != nil {
		return err
	}

	histIdx := 0
	migIdx := 0

	for histIdx < len(history) && migIdx < len(migs) {
		hist := history[histIdx]
		mig := migs[migIdx]

		if hist.Name != mig {
			return newError(nil, "mismatch between history %s and migration %s", hist.Name, mig)
		}

		content, err := dv.migs.Content(mig)
		if err != nil {
			return err
		}

		cks := checksum(content)
		if hist.Checksum != cks {
			return newError(nil, "checksum mismatch between history %s and migration %s", hist.Name, mig)
		}

		dv.logger.Log(
			fmt.Sprintf("Migration %s already done on %v",
				hist.Name,
				hist.Start,
			),
		)

		histIdx++
		migIdx++
	}

	if histIdx < len(history) {
		return newError(nil, "failed to find migration %v", history[histIdx])
	}

	return dv.migrate(ctx, migs[migIdx:])
}

func (dv DejaVu) migrate(ctx context.Context, names []string) error {
	for _, name := range names {
		dv.logger.Log(fmt.Sprintf("Processing migration %v...", name))

		content, err := dv.migs.Content(name)
		if err != nil {
			return err
		}

		if err = dv.db.Migrate(ctx, name, content); err != nil {
			return err
		}

		dv.logger.Log(fmt.Sprintf("Migration %v successfully processed", name))
	}

	return nil
}

func (dv DejaVu) lock(ctx context.Context) (Lock, error) {
	start := dv.clock.Now()

	lck, err := NewLock()
	if err != nil {
		return lck, err
	}

	lck.since = start

	if dv.db.Lock(ctx, lck) {
		return lck, nil
	}

	ticker := time.NewTicker(dv.tick)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-ctx.Done():
			return lck, newError(ctx.Err(), "canceling lock acquisition")
		case sig := <-sigs:
			return lck, newError(nil, "aborting lock acquisition because of %v", sig)
		case <-ticker.C:
			lck.since = dv.clock.Now()

			if lck.since.Sub(start) > dv.timeout {
				return lck, newError(nil, "failed to acquired lock after %v", dv.timeout)
			}

			if dv.db.Lock(ctx, lck) {
				return lck, nil
			}
		}
	}
}

func checksum(s string) string {
	data := sha256.Sum256([]byte(s))

	return base64.RawStdEncoding.EncodeToString(data[:])
}
