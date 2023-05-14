package models

import "time"

type RequeuedInfo struct {
	QueueStatus  QueueStatus
	DeliverAfter time.Time
}
