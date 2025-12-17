package model

import (
	"time"
)

// WorkingTimeStatus defines the state of the checkinout event processing.
type WorkingTimeStatus string

const (
	StatusWorkingPending    WorkingTimeStatus = "PENDING"
	StatusWorkingProcessing WorkingTimeStatus = "PROCESSING"
	StatusWorkingCompleted  WorkingTimeStatus = "COMPLETED"
	StatusWorkingFailed     WorkingTimeStatus = "FAILED"
)

// EmailStatus defines the state of the email event processing.
type EmailStatus string

const (
	StatusEmailPending    EmailStatus = "PENDING"
	StatusEmailProcessing EmailStatus = "PROCESSING"
	StatusEmailCompleted  EmailStatus = "COMPLETED"
	StatusEmailFailed     EmailStatus = "FAILED"
)

type WorkingTime struct {
	ID              int64             `json:"id"`
	EmployeeID      string            `json:"employeeId"`
	ClockInTime     time.Time         `json:"clockInTime"`
	ClockOutTime    *time.Time        `json:"clockOutTime,omitempty"`
	HoursWorked     float64           `json:"hoursWorked,omitempty"`
	RetryCount      int               `json:"retryCount"`
	LaborStatus     WorkingTimeStatus `json:"laborStatus"`
	EmailStatus     EmailStatus       `json:"emailStatus"`
	LaborRetryCount int               `json:"laborRetryCount"`
	EmailRetryCount int               `json:"emailRetryCount"`
}
