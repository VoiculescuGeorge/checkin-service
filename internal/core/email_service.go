package core

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailService interface {
	SendCheckOutSummary(ctx context.Context, to string, hours float64) error
}

type SESEmailService struct {
	client *ses.Client
	sender string
}

func NewSESEmailService(client *ses.Client, sender string) *SESEmailService {
	return &SESEmailService{client: client, sender: sender}
}

func (s *SESEmailService) SendCheckOutSummary(ctx context.Context, to string, hours float64) error {
	input := &ses.SendEmailInput{
		Source: aws.String(s.sender),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String("Work Shift Summary"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(fmt.Sprintf("Hello,\n\nYou have successfully checked out. Total hours worked: %.2f hours.", hours)),
				},
			},
		},
	}

	_, err := s.client.SendEmail(ctx, input)
	return err
}
