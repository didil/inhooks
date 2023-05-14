package models

import "time"

type QueuedInfo struct {
	MessageID    string
	QueueStatus  QueueStatus
	DeliverAfter time.Time
}
