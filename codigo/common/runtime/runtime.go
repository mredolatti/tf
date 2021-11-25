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

func NewImpl(logger log.Interface) (*Impl, error) {
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
}

func (i *Impl) Unblock() {
	select {
	case i.blocker <- struct{}{}:
	default:
		i.logger.Info("Finalizando ejecucion de servidor de indexado.")
	}
}

var _ Interface = (*Impl)(nil)
