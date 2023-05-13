package supervisor

import (
	"context"
	"sync"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/services"
	"go.uber.org/zap"
)

type Supervisor struct {
	logger               *zap.Logger
	messageFetcher       services.MessageFetcher
	ctx                  context.Context
	cancel               context.CancelFunc
	appConf              *lib.AppConfig
	inhooksConfigSvc     services.InhooksConfigService
	messageProcessor     services.MessageProcessor
	processingResultsSvc services.ProcessingResultsService
}

type SupervisorOpt func(s *Supervisor)

func NewSupervisor(opts ...SupervisorOpt) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Supervisor{}
	s.ctx = ctx
	s.cancel = cancel

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

func WithMessageFetcher(messageFetcher services.MessageFetcher) SupervisorOpt {
	return func(s *Supervisor) {
		s.messageFetcher = messageFetcher
	}
}

func WithAppConfig(appConf *lib.AppConfig) SupervisorOpt {
	return func(s *Supervisor) {
		s.appConf = appConf
	}
}

func WithInhooksConfigService(inhooksConfigSvc services.InhooksConfigService) SupervisorOpt {
	return func(s *Supervisor) {
		s.inhooksConfigSvc = inhooksConfigSvc
	}
}

func WithMessageProcessor(messageProcessor services.MessageProcessor) SupervisorOpt {
	return func(s *Supervisor) {
		s.messageProcessor = messageProcessor
	}
}

func WithProcessingResultsService(processingResultsSvc services.ProcessingResultsService) SupervisorOpt {
	return func(s *Supervisor) {
		s.processingResultsSvc = processingResultsSvc
	}
}

func (s *Supervisor) Start() {
	wg := &sync.WaitGroup{}
	flows := s.inhooksConfigSvc.GetFlows()
	for id := range flows {
		f := flows[id]

		for j := 0; j < len(f.Sinks); j++ {
			sink := f.Sinks[j]
			wg.Add(1)

			go func() {
				//TODO: move all from processing to ready (in case of previous crash)
				s.HandleReadyQueue(s.ctx, f, sink)
				wg.Done()
			}()
		}
	}

	wg.Wait()
}

func (s *Supervisor) Shutdown() {
	s.cancel()
}
