package log

import (
	"fmt"
	"io"
	"log"
)

const (
	None = iota
	Error
	Warning
	Info
	Debug
	Verbose
)

type Interface interface {
	Error(tpl string, params ...interface{})
	Warning(tpl string, params ...interface{})
	Info(tpl string, params ...interface{})
	Debug(tpl string, params ...interface{})
	Verbose(tpl string, params ...interface{})
}

type Impl struct {
	errLogger   log.Logger
	warnLogger  log.Logger
	infoLogger  log.Logger
	debugLogger log.Logger
	verLogger   log.Logger
	level       int
}

func New(w io.Writer, level int) (*Impl, error) {
	if level < None || level > Verbose {
		return nil, fmt.Errorf("unknown level: %d", level)
	}

	flags := log.Ldate | log.LUTC | log.Lshortfile
	return &Impl{
		level:       level,
		errLogger:   *log.New(w, "ERROR - ", flags),
		warnLogger:  *log.New(w, "WARNING - ", flags),
		infoLogger:  *log.New(w, "INFO - ", flags),
		debugLogger: *log.New(w, "DEBUG - ", flags),
		verLogger:   *log.New(w, "VERBOSE - ", flags),
	}, nil

}

func (i *Impl) Error(tpl string, params ...interface{}) {
	if i.level >= Error {
		i.errLogger.Output(3, fmt.Sprintf(tpl, params...))
	}
}

func (i *Impl) Warning(tpl string, params ...interface{}) {
	if i.level >= Warning {
		i.warnLogger.Output(3, fmt.Sprintf(tpl, params...))
	}
}

func (i *Impl) Info(tpl string, params ...interface{}) {
	if i.level >= Info {
		i.infoLogger.Output(3, fmt.Sprintf(tpl, params...))
	}
}

func (i *Impl) Debug(tpl string, params ...interface{}) {
	if i.level >= Debug {
		i.debugLogger.Output(3, fmt.Sprintf(tpl, params...))
	}
}

func (i *Impl) Verbose(tpl string, params ...interface{}) {
	if i.level >= Verbose {
		i.verLogger.Output(3, fmt.Sprintf(tpl, params...))
	}
}
