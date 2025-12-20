package core

import (
	"context"
	"errors"
	"time"

	"checkin.service/internal/core/model"
	"checkin.service/internal/ports/messaging"
	"checkin.service/internal/ports/repository"
)

type CheckInService struct {
	repo     repository.Repository
	producer messaging.SQSProducer
}

// NewCheckInService creates a new instance of our main application service,
// wiring up the database repository and the message queue producer.
func NewCheckInService(repo repository.Repository, p messaging.SQSProducer) *CheckInService {
	return &CheckInService{
		repo:     repo,
		producer: p,
	}
}

// ProcessCheckInOut is the core business logic. It figures out if an employee
// is clocking in or out by checking for an open work record.
func (s *CheckInService) ProcessCheckInOut(ctx context.Context, employeeID string) error {
	currentTime := time.Now().UTC()

	openWorkTime, err := s.repo.FindLastCheckIn(ctx, employeeID)
	if err != nil {
		return errors.New("failed to query last check-in")
	}

	if openWorkTime == nil {
		return s.handleCheckIn(ctx, employeeID, currentTime)
	}

	return s.handleCheckOut(ctx, openWorkTime, currentTime)
}

// UpdateWorkingTimeStatus is a simple pass-through to the repository layer,
// mainly used by background workers to update the status of a job.
func (s *CheckInService) UpdateWorkingTimeStatus(ctx context.Context, id int64, status model.WorkingTimeStatus, retryCount int) error {
	return s.repo.UpdateLaborStatus(ctx, id, status, retryCount)
}

// handleCheckIn handles the clock-in workflow.
func (s *CheckInService) handleCheckIn(ctx context.Context, employeeID string, clockIn time.Time) error {
	_, err := s.repo.CreateCheckIn(ctx, employeeID, clockIn)
	if err != nil {
		return errors.New("failed to create check-in record")
	}

	return nil
}

// handleCheckOut handles the clock-out workflow, including asynchronous work triggering.
func (s *CheckInService) handleCheckOut(ctx context.Context, workTime *model.WorkingTime, clockOut time.Time) error {
	duration := clockOut.Sub(workTime.ClockInTime)
	hoursWorked := duration.Hours()

	err := s.repo.UpdateCheckOut(ctx, workTime.ID, clockOut, hoursWorked, workTime.EmployeeID)
	if err != nil {
		return errors.New("failed to update check-out record")
	}

	emailEvent := messaging.EmailEvent{
		WorkingTimeID: workTime.ID,
		EmployeeID:    workTime.EmployeeID,
		HoursWorked:   hoursWorked,
		OccurredAt:    time.Now(),
	}
	s.producer.PublishEmail(ctx, emailEvent)

	checkInOutEvent := messaging.CheckOutEvent{
		WorkingTimeID: workTime.ID,
		EmployeeID:    workTime.EmployeeID,
		HoursWorked:   hoursWorked,
		ClockOutTime:  clockOut,
	}

	err = s.producer.PublishLabor(ctx, checkInOutEvent)

	if err != nil {
		return errors.New("failed to publish check-out event to queue")
	}

	return nil
}
