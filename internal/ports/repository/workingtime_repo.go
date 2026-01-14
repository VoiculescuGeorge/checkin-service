package repository

import (
	"context"
	"database/sql"
	"time"

	"checkin.service/internal/core/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

// WorkingTimeRepository is the concrete implementation for a PostgreSQL database.
type WorkingTimeRepository struct {
	DB *sql.DB
}

// NewWorkingTimeRepository create new instance
func NewWorkingTimeRepository(db *sql.DB) Repository {
	return &WorkingTimeRepository{DB: db}
}

// CreateCheckIn create checkin.
func (r *WorkingTimeRepository) CreateCheckIn(ctx context.Context, employeeID string, clockIn time.Time) (int64, error) {

	trace.SpanFromContext(ctx).SetAttributes(attribute.String("app.employeeId", employeeID))
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("app.employee_id", employeeID))

	var id int64
	query := `INSERT INTO working_times (employee_id, clock_in_time, labor_status, labor_retry_count, email_status, email_retry_count) 
              VALUES ($1, $2, $3, 0, $4, 0) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, employeeID, clockIn, model.StatusWorkingPending, model.StatusEmailPending).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateCheckOut do checkout.
func (r *WorkingTimeRepository) UpdateCheckOut(ctx context.Context, id int64, clockOut time.Time, hoursWorked float64, employeeID string) error {
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("app.employeeId", employeeID))
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("app.employee_id", employeeID))
	query := `UPDATE working_times 
              SET clock_out_time = $1, 
                  hours_worked = $2, 
                  labor_status = $3
              WHERE id = $4`

	_, err := r.DB.ExecContext(ctx, query, clockOut, hoursWorked, model.StatusWorkingPending, id)

	return err
}

// UpdateLaborStatus updates the status and retry count for a labor-related job.
func (r *WorkingTimeRepository) UpdateLaborStatus(ctx context.Context, id int64, status model.WorkingTimeStatus, retryCount int) error {

	query := `UPDATE working_times 
              SET labor_status = $1, 
                  labor_retry_count = $2 
              WHERE id = $3`

	_, err := r.DB.ExecContext(ctx, query, status, retryCount, id)

	return err
}

// FindLastCheckIn get last check in for a employee
func (r *WorkingTimeRepository) FindLastCheckIn(ctx context.Context, employeeID string) (*model.WorkingTime, error) {

	trace.SpanFromContext(ctx).SetAttributes(attribute.String("app.employeeId", employeeID))
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("app.employee_id", employeeID))

	var clockIn time.Time
	wt := &model.WorkingTime{EmployeeID: employeeID}

	query := `SELECT id, clock_in_time, labor_status, labor_retry_count
              FROM working_times
              WHERE employee_id = $1 AND clock_out_time IS NULL
              ORDER BY clock_in_time DESC
              LIMIT 1`

	row := r.DB.QueryRowContext(ctx, query, employeeID)
	err := row.Scan(&wt.ID, &clockIn, &wt.LaborStatus, &wt.LaborRetryCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	wt.ClockInTime = clockIn
	return wt, nil
}

// GetStatus retrieves just the status of a specific work record.
func (r *WorkingTimeRepository) GetStatus(ctx context.Context, id int64) (model.WorkingTimeStatus, error) {
	var status model.WorkingTimeStatus
	query := `SELECT labor_status FROM working_times WHERE id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(&status)
	if err != nil {
		return "", err
	}

	return status, nil
}

// GetCheckInOut fetches a complete working_times record by its ID.
func (r *WorkingTimeRepository) GetCheckInOut(ctx context.Context, id int64) (*model.WorkingTime, error) {
	query := `SELECT id, employee_id, labor_status, labor_retry_count, email_status, email_retry_count, hours_worked 
	          FROM working_times WHERE id = $1`

	wt := &model.WorkingTime{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&wt.ID, &wt.EmployeeID, &wt.LaborStatus, &wt.LaborRetryCount, &wt.EmailStatus, &wt.EmailRetryCount, &wt.HoursWorked,
	)
	if err != nil {
		return nil, err
	}
	return wt, nil
}

// UpdateEmailStatus updates the status and retry count for an email-related job.
func (r *WorkingTimeRepository) UpdateEmailStatus(ctx context.Context, id int64, status model.EmailStatus, retryCount int) error {
	query := `UPDATE working_times SET email_status = $1, email_retry_count = $2 WHERE id = $3`
	_, err := r.DB.ExecContext(ctx, query, status, retryCount, id)
	return err
}
