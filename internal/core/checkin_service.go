package checkin_service

import (
	"context"
	"time"

	"checkin.service/internal/core/model"
	"checkin.service/internal/ports/repository"
)

type CheckInService struct {
	repo repository.Repository
}

func NewCheckInService(repo repository.Repository) *CheckInService {
	return &CheckInService{
		repo: repo,
	}
}

func (s *CheckInService) CheckInOut(ctx context.Context, employeeID string) (*model.WorkingTime, error) {

	lastCheckIn, err := s.repo.FindLastCheckIn(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	if lastCheckIn == nil {
		clockInTime := time.Now()
		id, err := s.repo.CreateCheckIn(ctx, employeeID, clockInTime)
		if err != nil {
			return nil, err
		}

		return &model.WorkingTime{
			ID:          id,
			EmployeeID:  employeeID,
			ClockInTime: clockInTime,
			Status:      model.StatusPending,
		}, nil
	}

	clockOutTime := time.Now()
	hoursWorked := clockOutTime.Sub(lastCheckIn.ClockInTime).Hours()

	err = s.repo.UpdateCheckOut(ctx, lastCheckIn.ID, clockOutTime, hoursWorked)
	if err != nil {
		return nil, err
	}

	lastCheckIn.ClockOutTime = &clockOutTime
	lastCheckIn.HoursWorked = hoursWorked
	lastCheckIn.Status = model.StatusPending
	return lastCheckIn, nil
}

func (s *CheckInService) UpdateWorkingTimeStatus(ctx context.Context, id int64, status model.WorkingTimeStatus, retryCount int) error {
	return s.repo.UpdateStatus(ctx, id, status, retryCount)
}
