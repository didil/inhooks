package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/server"
	"github.com/didil/inhooks/pkg/server/handlers"
	"github.com/didil/inhooks/pkg/services"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	err := lib.LoadEnv()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	appConf, err := lib.ProcessAppConfig(ctx)
	if err != nil {
		log.Fatalf("failed to process config: %v", err)
	}

	logger, err := lib.NewLogger(appConf)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	inhooksConfigSvc := services.NewInhooksConfigService()
	err = inhooksConfigSvc.Load("inhooks.yml")
	if err != nil {
		logger.Fatal("failed to load inhooks config", zap.Error(err))
	}

	messageDecoder := services.NewMessageDecoder()

	app := handlers.NewApp(
		handlers.WithLogger(logger),
		handlers.WithAppConfig(appConf),
		handlers.WithInhooksConfigService(inhooksConfigSvc),
		handlers.WithMessageDecoder(messageDecoder),
	)

	r := server.NewRouter(app)

	addr := fmt.Sprintf("%s:%d", appConf.Server.Host, appConf.Server.Port)

	logger.Info("Listening ...", zap.String("addr", addr))
	err = http.ListenAndServe(addr, r)
	if err != nil {
		logger.Fatal("listener failure", zap.Error(err))
	}
}
