package labor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"checkin.service/internal/core/model"
	"checkin.service/internal/ports/messaging"
	"checkin.service/internal/ports/repository"
	legacyAPI "checkin.service/internal/worker/legacyapi"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sony/gobreaker"
)

// LaberProcessor handles jobs from the labor queue, which involves calling a legacy API.
// It uses a circuit breaker to avoid hammering the legacy system if it's having issues.
type LaberProcessor struct {
	Repo      repository.Repository
	legacyapi legacyAPI.LegacyAPIClient
	cb        *gobreaker.CircuitBreaker
}

// NewProcessor creates a new processor for the labor queue. It sets up a
// circuit breaker to protect the legacy API from being overwhelmed.
func NewProcessor(r repository.Repository, legacyapi legacyAPI.LegacyAPIClient) *LaberProcessor {
	settings := gobreaker.Settings{
		Name:        "Legacy-API",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if failure rate is bigger then 50% after at least 10 requests
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.5
		},
	}

	return &LaberProcessor{
		Repo:      r,
		legacyapi: legacyapi,
		cb:        gobreaker.NewCircuitBreaker(settings),
	}
}

// Process is the core logic for handling a message from the labor queue.
// It calls the legacy API through a circuit breaker and handles retries with exponential backoff.
func (p *LaberProcessor) Process(ctx context.Context, msg types.Message) (bool, int32, error) {
	var event messaging.CheckOutEvent
	if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
		log.Printf("Failed to unmarshal labor event: %v", err)
		return false, 0, err // Do not retry on malformed message
	}

	log.Printf("Processing check-out for Employee: %s, Hours: %.2f", event.EmployeeID, event.HoursWorked)

	record, err := p.Repo.GetCheckInOut(ctx, event.WorkingTimeID)
	if err != nil {
		return true, 10, fmt.Errorf("failed to get record from db: %w", err)
	}

	if record.LaborStatus == model.StatusWorkingCompleted {
		return false, 0, nil
	}

	_, err = p.cb.Execute(func() (interface{}, error) {
		return nil, p.legacyapi.RecordCheckOut(ctx, event)
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			fmt.Println("Circuit Breaker is OPEN; skipping Legacy API call")
		}
		newCount := record.RetryCount + 1
		p.Repo.UpdateLaborStatus(ctx, event.WorkingTimeID, model.StatusWorkingPending, newCount)

		delay := calculateBackoff(newCount)
		return true, delay, err
	}

	err = p.Repo.UpdateLaborStatus(ctx, event.WorkingTimeID, model.StatusWorkingCompleted, 0)
	return false, 0, err
}

// calculateBackoff determines how long to wait before retrying a failed job.
// It increases the delay exponentially with each retry.
func calculateBackoff(retryCount int) int32 {
	backoff := int32(math.Pow(2, float64(retryCount)) * 10)
	if backoff > 3600 {
		return 3600 // max at 1 hour
	}
	return backoff
}
