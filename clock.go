package dejavu

import (
	"fmt"
	"time"
)

type Clock interface {
	fmt.Stringer

	Now() time.Time
}

func NewUtcClock() UtcClock {
	return UtcClock{}
}

type UtcClock struct{}

func (c UtcClock) Now() time.Time {
	return time.Now().UTC()
}

func (c UtcClock) String() string {
	return "UTC Clock"
}
