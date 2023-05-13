package models

type QueueStatus string

const (
	QueueStatusScheduled  QueueStatus = "scheduled"
	QueueStatusReady      QueueStatus = "ready"
	QueueStatusProcessing QueueStatus = "processing"
	QueueStatusDone       QueueStatus = "done"
	QueueStatusDead       QueueStatus = "dead"
)
