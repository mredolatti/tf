package runtime

import (
	"errors"

	"github.com/mredolatti/tf/codigo/common/log"
)

var ErrNoLogger = errors.New("logger cannot be nil")

type Interface interface {
	Block()
	Unblock()
}

type Impl struct {
	logger  log.Interface
	blocker chan struct{}
}

func New(logger log.Interface) (*Impl, error) {
	if logger == nil {
		return nil, ErrNoLogger
	}

	return &Impl{
		logger:  logger,
		blocker: make(chan struct{}, 1),
	}, nil
}

func (i *Impl) Block() {
	<-i.blocker
	i.logger.Info("Finalizando ejecucion de servidor de la aplicacion.")
}

func (i *Impl) Unblock() {
	select {
	case i.blocker <- struct{}{}:
	default:
	}
}

var _ Interface = (*Impl)(nil)
