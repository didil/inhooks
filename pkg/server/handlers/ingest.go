package handlers

import (
	"fmt"
	"net/http"

	"github.com/didil/inhooks/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ingestRequestsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "ingest_requests_total",
	Help: "Number of ingest requests",
})

var enqueuedMessagesCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "enqueued_messages_total",
	Help: "Number of enqueued messages",
})

func (app *App) HandleIngest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	sourceSlug := chi.URLParam(r, "sourceSlug")
	logger := app.logger.With(zap.String("reqID", reqID), zap.String("sourceSlug", sourceSlug))

	logger.Info("new ingest request")
	ingestRequestsCounter.Inc()

	// find the flow
	flow := app.inhooksConfigSvc.FindFlowForSource(sourceSlug)
	if flow == nil {
		logger.Error("ingest request failed: unknown source slug", zap.String("sourceSlug", sourceSlug))
		app.WriteJSONErr(w, http.StatusNotFound, reqID, fmt.Errorf("unknown source slug %s", sourceSlug))
		return
	}

	logger = logger.With(zap.String("flowID", flow.ID), zap.String("sourceID", flow.Source.ID))

	// build messages
	messages, err := app.messageBuilder.FromHttp(flow, r, reqID)
	if err != nil {
		logger.Error("ingest request failed: unable to build messages", zap.Error(err))
		app.WriteJSONErr(w, http.StatusBadRequest, reqID, fmt.Errorf("unable to read data"))
		return
	}

	// verify messages (first message is enough as payloads and signatures are the same)
	err = app.messageVerifier.Verify(flow, messages[0])
	if err != nil {
		logger.Error("ingest request failed: unable to verify messages signature", zap.Error(err))
		app.WriteJSONErr(w, http.StatusForbidden, reqID, fmt.Errorf("unable to verify signature"))
		return
	}

	// enqueue messages
	queuedInfos, err := app.messageEnqueuer.Enqueue(ctx, messages)
	if err != nil {
		logger.Error("ingest request failed: unable to enqueue messages", zap.Error(err))
		app.WriteJSONErr(w, http.StatusBadRequest, reqID, fmt.Errorf("unable to enqueue data"))
		return
	}
	enqueuedMessagesCounter.Add(float64(len(queuedInfos)))

	for _, queuedInfo := range queuedInfos {
		fields := []zapcore.Field{zap.String("messageID", queuedInfo.MessageID), zap.String("queue", string(queuedInfo.QueueStatus))}
		if queuedInfo.QueueStatus == models.QueueStatusScheduled {
			fields = append(fields, zap.Time("nextAttemptAfter", queuedInfo.DeliverAfter))
		}
		logger.Info("message queued", fields...)
	}

	app.WriteJSONResponse(w, http.StatusOK, JSONOK{})
	logger.Info("ingest request succeeded")
}
