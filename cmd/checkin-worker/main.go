package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"checkin.service/internal/config"
	"checkin.service/internal/ports/repository"
	"checkin.service/internal/worker"
	"checkin.service/internal/worker/labor"
	legacyAPI "checkin.service/internal/worker/legacyapi"
	"checkin.service/pkg/aws"
	"checkin.service/pkg/database"
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
	awsCfg, err := aws.NewAWSConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	// Initialize Dependencies
	sqsClient := sqs.NewFromConfig(awsCfg)

	repo := repository.NewWorkingTimeRepository(db)

	legacyClient := legacyAPI.NewHTTPClient(cfg.LegacyAPIURL)
	processor := labor.NewProcessor(repo, legacyClient)

	// Start Worker
	ctx, cancel := context.WithCancel(context.Background())
	app := worker.NewWorker(sqsClient, cfg.LaborSQSQueueURL, processor)

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
