package resolver

import "log"

type StdLogger struct{}

func (l *StdLogger) Info(msg string) {
	log.Println(msg)
}
