package dejavu

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/template"
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

func (dv DejaVu) Missing(ctx context.Context) ([]string, error) {
	history, err := dv.db.History(ctx)
	if err != nil {
		return nil, err
	}

	migs, err := dv.migs.List(dv.db.Name())
	if err != nil {
		return nil, err
	}

	histIdx := 0
	migIdx := 0

	for histIdx < len(history) && migIdx < len(migs) {
		hist := history[histIdx]
		mig := migs[migIdx]

		if hist.Name != mig {
			return nil, newError(nil, "mismatch between history %s and migration %s", hist.Name, mig)
		}

		content, err := dv.migs.Content(mig)
		if err != nil {
			return nil, err
		}

		cks := checksum(content)
		if hist.Checksum != cks {
			return nil, newError(nil, "checksum mismatch between history %s and migration %s", hist.Name, mig)
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
		return nil, newError(nil, "failed to find migration %v", history[histIdx])
	}

	return migs[migIdx:], nil
}

func (dv DejaVu) Upgrade(ctx context.Context) error {
	dv.logger.Log("Starting database upgrade...")

	if err := dv.db.Init(ctx); err != nil {
		return err
	}

	lck, err := dv.lock(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err2 := dv.db.Unlock(ctx, lck); err2 != nil {
			dv.logger.Log(fmt.Sprintf("failed to free lock: %v", err2))

			if err == nil {
				err = err2
			}
		}
	}()

	if err = dv.doUpgrade(ctx); err != nil {
		return err
	}

	dv.logger.Log("Database successfully upgraded")

	return err
}

func (dv DejaVu) doUpgrade(ctx context.Context) error {
	migs, err := dv.Missing(ctx)
	if err != nil {
		return err
	}

	for _, mig := range migs {
		dv.logger.Log(fmt.Sprintf("Processing migration %v...", mig))

		content, err := dv.migs.Content(mig)
		if err != nil {
			return err
		}

		tmpl, err := template.New(mig).Parse(content)
		if err != nil {
			return newError(err, "failed to parse template %s", mig)
		}

		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, nil); err != nil {
			return newError(err, "failed to execute template %s", mig)
		}

		if err = dv.db.Migrate(ctx, mig, buf.String()); err != nil {
			return err
		}

		dv.logger.Log(fmt.Sprintf("Migration %v successfully processed", mig))
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
