package entity

import "time"

type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusRunning   JobStatus = "RUNNING"
	StatusFailed    JobStatus = "FAILED"
	StatusCompleted JobStatus = "COMPLETED"
)

type Job struct {
	ID          string    `json:"id"`
	Task        string    `json:"task"`
	Token       *string   `json:"token,omitempty"` // untuk idempotency
	Status      JobStatus `json:"status"`
	Attempts    int32     `json:"attempts"`
	MaxAttempts int32     `json:"max_attempts"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JobStatusSummary struct {
	Pending   int32 `json:"pending"`
	Running   int32 `json:"running"`
	Failed    int32 `json:"failed"`
	Completed int32 `json:"completed"`
}
