package dejavu

import (
	"fmt"
	"os"
	"time"
)

const (
	lockID = 1
)

type Lock struct {
	id       int
	hostname string
	pid      int
	since    time.Time
}

func NewLock() (Lock, error) {
	result := Lock{
		id:  lockID,
		pid: os.Getpid(),
	}

	hostname, err := os.Hostname()
	if err != nil {
		return result, err
	}

	result.hostname = hostname

	return result, nil
}

func (l Lock) String() string {
	return fmt.Sprintf("Lock from host %s by PID %d since %v", l.hostname, l.pid, l.since)
}
