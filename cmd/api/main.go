package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
		handlers.WithAppConfig(appConf),
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

	go func() {
		logger.Info("listening ...", zap.String("addr", addr))
		err = httpServer.ListenAndServe()
		if err != nil {
			logger.Fatal("listener failure", zap.Error(err))
		}
	}()

	httpClient := lib.NewHttpClient(appConf)

	messageProcessor := services.NewMessageProcessor(httpClient)
	processingResultsService := services.NewProcessingResultsService(timeSvc, redisStore)

	svisor := supervisor.NewSupervisor(
		supervisor.WithLogger(logger),
		supervisor.WithMessageFetcher(messageFetcher),
		supervisor.WithAppConfig(appConf),
		supervisor.WithInhooksConfigService(inhooksConfigSvc),
		supervisor.WithMessageProcessor(messageProcessor),
		supervisor.WithProcessingResultsService(processingResultsService),
	)

	go func() {
		logger.Info("starting supervisor ...")
		svisor.Start()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	svisor.Shutdown()

	serverShutdownContext, cancel := context.WithTimeout(context.Background(), appConf.Server.ShutdownGracePeriod)
	defer cancel()
	err = httpServer.Shutdown(serverShutdownContext)
	if err != nil {
		logger.Fatal("http server shutdown failed", zap.Error(err))
	}
}
