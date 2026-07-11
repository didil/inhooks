package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var DecisionsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rate_limit_decisions_total",
		Help: "Total rate-limit decisions per (flow, sink).",
	},
	[]string{"flow_id", "sink_id", "decision"},
)

var TokensRemaining = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "rate_limit_tokens_remaining",
		Help: "Current token-bucket remaining tokens (sampled).",
	},
	[]string{"flow_id", "sink_id"},
)
