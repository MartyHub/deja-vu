package dejavu

import (
	"fmt"
	"time"
)

type Migration struct {
	Name       string
	Start      time.Time
	DurationMs int64
	Checksum   string
}

func (m Migration) String() string {
	return fmt.Sprintf("Migration %s started at %v last %d", m.Name, m.Start, m.DurationMs)
}
