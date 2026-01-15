package email

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"checkin.service/internal/core/model"
	core "checkin.service/internal/core/service"
	"checkin.service/internal/ports/messaging"
	"checkin.service/internal/ports/repository"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

type EmailProcessor struct {
	emailService core.EmailService
	repo         repository.Repository
}

// NewProcessor sets up a new processor for handling email-related jobs.
// It needs an email service to send emails and a repository to update the job status.
func NewProcessor(emailService core.EmailService, repo repository.Repository) *EmailProcessor {
	return &EmailProcessor{
		emailService: emailService,
		repo:         repo,
	}
}

// Process is the main entry point for handling a message from the email queue.
// It tries to send an email and will tell the worker to retry if something goes wrong.
func (p *EmailProcessor) Process(ctx context.Context, msg types.Message) (bool, int32, error) {
	var event messaging.EmailEvent
	if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal email event")
		return false, 0, err // Do not retry on malformed message
	}

	record, err := p.repo.GetCheckInOut(ctx, event.WorkingTimeID)
	if err != nil {
		// If we can't get the record, retry after a short delay.
		return true, 10, fmt.Errorf("failed to get record from db for email processing: %w", err)
	}

	if record.EmailStatus == model.StatusEmailCompleted {
		log.Ctx(ctx).Info().Int64("working_time_id", event.WorkingTimeID).Msg("Email already sent. Skipping.")
		return false, 0, nil
	}

	err = p.emailService.SendCheckOutSummary(ctx, event.EmployeeID+"@factory.com", event.HoursWorked)
	if err != nil {
		newCount := record.EmailRetryCount + 1
		p.repo.UpdateEmailStatus(ctx, event.WorkingTimeID, model.StatusEmailPending, newCount)

		delay := calculateBackoff(newCount)
		return true, delay, err
	}

	err = p.repo.UpdateEmailStatus(ctx, event.WorkingTimeID, model.StatusEmailCompleted, 0)
	return false, 0, err
}

// calculateBackoff determines how long to wait before retrying a failed job.
// It increases the delay exponentially with each retry to avoid overwhelming a struggling service.
func calculateBackoff(retryCount int) int32 {
	backoff := int32(math.Pow(2, float64(retryCount)) * 10)
	if backoff > 3600 { // Cap at 1 hour
		return 3600
	}
	return backoff
}
