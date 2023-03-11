package dejavu

import (
	"fmt"
	"log"
)

type Logger interface {
	fmt.Stringer

	Log(string)
}

type FmtLogger struct{}

func (l FmtLogger) Log(s string) {
	fmt.Println(s) //nolint:forbidigo
}

func (l FmtLogger) String() string {
	return "fmt logger"
}

type LogLogger struct{}

func (l LogLogger) Log(s string) {
	log.Println(s)
}

func (l LogLogger) String() string {
	return "log logger"
}
