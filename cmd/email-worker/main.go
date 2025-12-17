package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"checkin.service/internal/config"
	"checkin.service/internal/core"
	"checkin.service/internal/ports/repository"
	"checkin.service/internal/worker"
	"checkin.service/internal/worker/email"
	"checkin.service/pkg/database"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// DB connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to the database.")

	// AWS SDK Config
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	// Initialize Dependencies
	sqsClient := sqs.NewFromConfig(awsCfg)
	sesClient := ses.NewFromConfig(awsCfg)
	repo := repository.NewWorkingTimeRepository(db)
	emailService := core.NewSESEmailService(sesClient, "checkOut@checkout-service.com")
	processor := email.NewProcessor(emailService, repo)

	// Start Worker
	ctx, cancel := context.WithCancel(context.Background())
	app := worker.NewWorker(sqsClient, cfg.EmailSQSQueueURL, processor)

	go func() {
		app.Start(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down worker...")

	// Cancel the context to signal the worker to stop polling.
	cancel()

	log.Println("Worker exited gracefully")
}
