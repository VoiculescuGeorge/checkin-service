package model

import (
	"time"
)

// WorkingTimeStatus defines the state of the checkout event processing.
type WorkingTimeStatus string

const (
	StatusPending    WorkingTimeStatus = "PENDING"
	StatusProcessing WorkingTimeStatus = "PROCESSING"
	StatusCompleted  WorkingTimeStatus = "COMPLETED"
	StatusFailed     WorkingTimeStatus = "FAILED"
)

type WorkingTime struct {
	ID           int64             `json:"id"`
	EmployeeID   string            `json:"employeeId"`
	ClockInTime  time.Time         `json:"clockInTime"`
	ClockOutTime *time.Time        `json:"clockOutTime,omitempty"`
	HoursWorked  float64           `json:"hoursWorked,omitempty"`
	Status       WorkingTimeStatus `json:"status"`
	RetryCount   int               `json:"retryCount"`
}
