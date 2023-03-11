package dejavu

import "testing"

type testLogger struct {
	t *testing.T
}

func newTestLogger(t *testing.T) testLogger {
	t.Helper()

	return testLogger{t: t}
}

func (l testLogger) Log(s string) {
	l.t.Log(s)
}

func (l testLogger) String() string {
	return "test logger"
}
