package models

import (
	"net/http"
	"time"
)

type Message struct {
	ID          string      `json:"id"`
	FlowID      string      `json:"flowID"`
	SourceID    string      `json:"sourceID"`
	SinkID      string      `json:"sinkID"`
	HttpHeaders http.Header `json:"httpHeaders"`
	RawQuery    string      `json:"rawQuery"`
	Payload     []byte      `json:"payload"`

	// Processing Info
	DeliveryAttempts []*DeliveryAttempt `json:"deliveryAttempts"`
	DeliverAfter     time.Time
}

type DeliveryAttempt struct {
	At     time.Time             `json:"at"`
	Status DeliveryAttemptStatus `json:"status"`
}

type DeliveryAttemptStatus string

const (
	DeliveryAttemptStatusOK    = "ok"
	DeliveryAttemptStatusailed = "failed"
)

type MessageStatus string
