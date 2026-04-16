package logger

import (
	"log"
	"os"
)

type Logger interface {
	Errorf(format string, args ...any)
	Errorln(args ...any)
	Infof(format string, args ...any)
	Infoln(args ...any)
}

type ConsoleLogger struct {
	Error *log.Logger
	Info  *log.Logger
}

func NewConsoleLogger() (*ConsoleLogger, error) {
	return &ConsoleLogger{
		Error: log.New(os.Stderr, "ERROR\t", log.Flags()|log.LUTC),
		Info:  log.New(os.Stdout, "INFO\t", log.Flags()|log.LUTC),
	}, nil
}

func (cl *ConsoleLogger) Errorf(format string, args ...any) {
	cl.Error.Printf(format, args...)
}

func (cl *ConsoleLogger) Errorln(args ...any) {
	cl.Error.Println(args...)
}

func (cl *ConsoleLogger) Infof(format string, args ...any) {
	cl.Info.Printf(format, args...)
}

func (cl *ConsoleLogger) Infoln(args ...any) {
	cl.Info.Println(args...)

}
