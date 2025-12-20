package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"checkin.service/internal/config"
	"checkin.service/internal/core"
	"checkin.service/internal/ports/repository"
	"checkin.service/internal/worker"
	"checkin.service/internal/worker/email"
	"checkin.service/pkg/aws"
	"checkin.service/pkg/database"
	"checkin.service/pkg/logger"
	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not load configuration")
	}

	// Configure structured logging
	logger.Setup(cfg.IsLocalDev)

	// Configure OpenTelemetry Tracing
	shutdownTracer, err := telemetry.InitTracer("email-worker")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init tracer")
	}
	defer func() {
		_ = shutdownTracer(context.Background())
	}()

	// DB connection
	db, err := database.NewInstrumentedConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening database")
	}
	defer db.Close()
	log.Info().Msg("Successfully connected to the database.")

	// AWS SDK Config
	awsCfg, err := aws.NewAWSConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load SDK config")
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
	log.Info().Msg("Shutting down worker...")

	// Cancel the context to signal the worker to stop polling.
	cancel()

	log.Info().Msg("Worker exited gracefully")
}
