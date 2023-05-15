package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/server"
	"github.com/didil/inhooks/pkg/server/handlers"
	"github.com/didil/inhooks/pkg/services"
	"github.com/didil/inhooks/pkg/supervisor"
	versionpkg "github.com/didil/inhooks/pkg/version"
	"go.uber.org/zap"
)

var (
	version = "dev"
)

func main() {
	versionpkg.SetVersion(version)

	err := lib.LoadEnv()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	appConf, err := lib.InitAppConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to process config: %v", err)
	}

	logger, err := lib.NewLogger(appConf)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	inhooksConfigSvc := services.NewInhooksConfigService(logger, appConf)
	logger.Info("loading inhooks config", zap.String("inhooksConfigFile", appConf.InhooksConfigFile))

	err = inhooksConfigSvc.Load(appConf.InhooksConfigFile)
	if err != nil {
		logger.Fatal("failed to load inhooks config", zap.Error(err))
	}

	timeSvc := services.NewTimeService()

	messageBuilder := services.NewMessageBuilder(timeSvc)

	redisClient, err := lib.InitRedisClient(appConf)
	if err != nil {
		logger.Fatal("failed to init redis client", zap.Error(err))
	}
	redisStore, err := services.NewRedisStore(redisClient, appConf.Redis.InhooksDBName)
	if err != nil {
		logger.Fatal("failed to init redis store", zap.Error(err))
	}

	messageEnqueuer := services.NewMessageEnqueuer(redisStore, timeSvc)
	messageFetcher := services.NewMessageFetcher(redisStore, timeSvc)

	app := handlers.NewApp(
		handlers.WithLogger(logger),
		handlers.WithInhooksConfigService(inhooksConfigSvc),
		handlers.WithMessageBuilder(messageBuilder),
		handlers.WithMessageEnqueuer(messageEnqueuer),
	)

	r := server.NewRouter(app)

	addr := fmt.Sprintf("%s:%d", appConf.Server.Host, appConf.Server.Port)
	httpServer := http.Server{
		Addr:    addr,
		Handler: r,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		logger.Info("listening ...", zap.String("addr", addr))
		err = httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("listener failure", zap.Error(err))
		}
		logger.Info("http server shutdown")
		wg.Done()
	}()

	httpClient := lib.NewHttpClient(appConf)

	messageProcessor := services.NewMessageProcessor(httpClient)
	retryCalculator := services.NewRetryCalculator()
	processingResultsSvc := services.NewProcessingResultsService(timeSvc, redisStore, retryCalculator)
	schedulerSvc := services.NewSchedulerService(redisStore, timeSvc)

	processingRecoverySvc, err := services.NewProcessingRecoveryService(redisStore)
	if err != nil {
		logger.Fatal("failed to init ProcessingRecoveryService", zap.Error(err))
	}

	svisor := supervisor.NewSupervisor(
		supervisor.WithLogger(logger),
		supervisor.WithMessageFetcher(messageFetcher),
		supervisor.WithAppConfig(appConf),
		supervisor.WithInhooksConfigService(inhooksConfigSvc),
		supervisor.WithMessageProcessor(messageProcessor),
		supervisor.WithProcessingResultsService(processingResultsSvc),
		supervisor.WithSchedulerService(schedulerSvc),
		supervisor.WithProcessingRecoveryService(processingRecoverySvc),
	)

	wg.Add(1)
	go func() {
		logger.Info("starting supervisor ...")
		svisor.Start()
		logger.Info("supervisor shutdown")
		wg.Done()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	sig := <-sigs
	logger.Info("received shutdown signal, shutting down process", zap.String("signal", sig.String()))

	svisor.Shutdown()

	serverShutdownContext, cancel := context.WithTimeout(context.Background(), appConf.Server.ShutdownGracePeriod)
	defer cancel()
	err = httpServer.Shutdown(serverShutdownContext)
	if err != nil {
		logger.Fatal("http server shutdown failed", zap.Error(err))
	}

	wg.Wait()
}
