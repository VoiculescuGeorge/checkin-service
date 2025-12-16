package repository

import (
	"context"
	"database/sql"
	"time"

	"checkin.service/internal/core/model"
)

// Repository is the interface defining the contract for data access.
type Repository interface {
	CreateCheckIn(ctx context.Context, employeeID string, clockIn time.Time) (int64, error)
	UpdateCheckOut(ctx context.Context, id int64, clockOut time.Time, hoursWorked float64) error
	UpdateStatus(ctx context.Context, id int64, status model.WorkingTimeStatus, retryCount int) error
	FindLastCheckIn(ctx context.Context, employeeID string) (*model.WorkingTime, error)
}

// WorkingTimeRepository is the concrete implementation for a PostgreSQL database.
type WorkingTimeRepository struct {
	DB *sql.DB
}

// NewWorkingTimeRepository creates a new instance of the repository.
func NewWorkingTimeRepository(db *sql.DB) Repository {
	return &WorkingTimeRepository{DB: db}
}

// CreateCheckIn implements the Repository interface.
func (r *WorkingTimeRepository) CreateCheckIn(ctx context.Context, employeeID string, clockIn time.Time) (int64, error) {

	var id int64
	query := `INSERT INTO working_times (employee_id, clock_in_time, status, retry_count) 
              VALUES ($1, $2, $3, 0) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, employeeID, clockIn, model.StatusPending).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateCheckOut implements the Repository interface.
func (r *WorkingTimeRepository) UpdateCheckOut(ctx context.Context, id int64, clockOut time.Time, hoursWorked float64) error {

	query := `UPDATE working_times 
              SET clock_out_time = $1, 
                  hours_worked = $2, 
                  status = $3
              WHERE id = $4`

	_, err := r.DB.ExecContext(ctx, query, clockOut, hoursWorked, model.StatusPending, id)

	return err
}

// UpdateStatus implements the Repository interface.
func (r *WorkingTimeRepository) UpdateStatus(ctx context.Context, id int64, status model.WorkingTimeStatus, retryCount int) error {

	query := `UPDATE working_times 
              SET status = $1, 
                  retry_count = $2 
              WHERE id = $3`

	_, err := r.DB.ExecContext(ctx, query, status, retryCount, id)

	return err
}

// FindLastCheckIn implements the Repository interface.
func (r *WorkingTimeRepository) FindLastCheckIn(ctx context.Context, employeeID string) (*model.WorkingTime, error) {

	var clockIn time.Time
	wt := &model.WorkingTime{EmployeeID: employeeID}

	query := `SELECT id, clock_in_time, status, retry_count
              FROM working_times
              WHERE employee_id = $1 AND clock_out_time IS NULL
              ORDER BY clock_in_time DESC
              LIMIT 1`

	row := r.DB.QueryRowContext(ctx, query, employeeID)
	err := row.Scan(&wt.ID, &clockIn, &wt.Status, &wt.RetryCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	wt.ClockInTime = clockIn
	return wt, nil
}
