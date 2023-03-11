package dejavu

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = time.UnixMilli(1678485867000) //nolint:gochecknoglobals

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func (c fixedClock) String() string {
	return fmt.Sprintf("Fixed clock at %v", c.now)
}

func newTestClock() fixedClock {
	return fixedClock{now: now}
}

func TestNewUtcClock(t *testing.T) {
	assert.NotNil(t, NewUtcClock())
}

func Test_utcClock_Now(t *testing.T) {
	now := time.Now().UTC()

	assert.GreaterOrEqual(t, NewUtcClock().Now(), now)
}

func Test_utcClock_String(t *testing.T) {
	assert.Equal(t, "UTC Clock", UtcClock{}.String())
}
