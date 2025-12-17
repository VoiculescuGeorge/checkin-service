package messaging

import "time"

// CheckOutEvent is the JSON payload sent via SQS for checkinout queue
type CheckOutEvent struct {
	WorkingTimeID int64     `json:"workingTimeId"`
	EmployeeID    string    `json:"employeeId"`
	HoursWorked   float64   `json:"hoursWorked"`
	ClockOutTime  time.Time `json:"clockOutTime"`
}

// EmailEvent is the JSON payload sent via SQS for email queue
type EmailEvent struct {
	WorkingTimeID int64     `json:"workingTimeId"`
	EmployeeID    string    `json:"employeeId"`
	HoursWorked   float64   `json:"hoursWorked"`
	OccurredAt    time.Time `json:"occurredAt"`
}