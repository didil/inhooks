package supervisor

import (
	"context"
	"sync"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/services"
	"go.uber.org/zap"
)

type Supervisor struct {
	logger                *zap.Logger
	messageFetcher        services.MessageFetcher
	ctx                   context.Context
	cancel                context.CancelFunc
	appConf               *lib.AppConfig
	inhooksConfigSvc      services.InhooksConfigService
	messageProcessor      services.MessageProcessor
	processingResultsSvc  services.ProcessingResultsService
	schedulerSvc          services.SchedulerService
	processingRecoverySvc services.ProcessingRecoveryService
	cleanupSvc            services.CleanupService
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

func WithSchedulerService(schedulerService services.SchedulerService) SupervisorOpt {
	return func(s *Supervisor) {
		s.schedulerSvc = schedulerService
	}
}

func WithProcessingRecoveryService(processingRecoverySvc services.ProcessingRecoveryService) SupervisorOpt {
	return func(s *Supervisor) {
		s.processingRecoverySvc = processingRecoverySvc
	}
}

func WithCleanupService(cleanupSvc services.CleanupService) SupervisorOpt {
	return func(s *Supervisor) {
		s.cleanupSvc = cleanupSvc
	}
}

func (s *Supervisor) Start() {
	wg := &sync.WaitGroup{}
	flows := s.inhooksConfigSvc.GetFlows()
	for id := range flows {
		f := flows[id]

		for j := 0; j < len(f.Sinks); j++ {
			sink := f.Sinks[j]
			logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))

			wg.Add(4)

			go func() {
				s.HandleProcessingQueue(f, sink)
				logger.Info("processing queue handler shutdown")
				wg.Done()
			}()

			go func() {
				s.HandleReadyQueue(f, sink)
				logger.Info("ready queue handler shutdown")
				wg.Done()
			}()

			go func() {
				s.HandleScheduledQueue(f, sink)
				logger.Info("scheduled queue handler shutdown")
				wg.Done()
			}()

			go func() {
				s.HandleDoneQueue(f, sink)
				logger.Info("done queue handler shutdown")
				wg.Done()
			}()
		}
	}

	wg.Wait()
}

func (s *Supervisor) Shutdown() {
	s.cancel()
}
