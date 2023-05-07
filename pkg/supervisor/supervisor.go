package supervisor

import (
	"time"

	"go.uber.org/zap"
)

type Supervisor struct {
	logger *zap.Logger
}

type SupervisorOpt func(s *Supervisor)

func NewSupervisor(opts ...SupervisorOpt) *Supervisor {
	s := &Supervisor{}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithLogger(logger *zap.Logger) SupervisorOpt {
	return func(s *Supervisor) {
		s.logger = logger
	}
}

const readyQueueCheckInterval = 1 * time.Second

func (s *Supervisor) Start() {

	go func() {
		// check if there are any messages in ready queue

		time.Sleep(readyQueueCheckInterval)
	}()

}
