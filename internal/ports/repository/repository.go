package repository

import (
	"context"
	"time"

	"checkin.service/internal/core/model"
)

// Repository contract
type Repository interface {
	GetCheckInOut(ctx context.Context, id int64) (*model.WorkingTime, error)
	CreateCheckIn(ctx context.Context, employeeID string, clockIn time.Time) (int64, error)
	UpdateCheckOut(ctx context.Context, id int64, clockOut time.Time, hoursWorked float64, employeeID string) error
	UpdateLaborStatus(ctx context.Context, id int64, status model.WorkingTimeStatus, retryCount int) error
	FindLastCheckIn(ctx context.Context, employeeID string) (*model.WorkingTime, error)
	GetStatus(ctx context.Context, id int64) (model.WorkingTimeStatus, error)
	UpdateEmailStatus(ctx context.Context, id int64, status model.EmailStatus, retryCount int) error
}
