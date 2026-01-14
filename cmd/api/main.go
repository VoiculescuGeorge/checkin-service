// Entry point for REST API
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"checkin.service/internal/api"
	"checkin.service/internal/config"
	checkin_service "checkin.service/internal/core"
	"checkin.service/internal/ports/messaging"
	"checkin.service/internal/ports/repository"
	"checkin.service/pkg/aws"
	"checkin.service/pkg/database"
	"checkin.service/pkg/logger"
	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	shutdownTracer, err := telemetry.InitTracer("checkin-api")
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

	// Initialize dependencies
	sqsClient := sqs.NewFromConfig(awsCfg)
	repo := repository.NewWorkingTimeRepository(db)
	producer := messaging.NewSQSProducer(sqsClient, cfg.LaborSQSQueueURL, cfg.EmailSQSQueueURL)
	coreService := checkin_service.NewCheckInService(repo, producer)

	// Setup router and server
	router := api.NewRouter(*coreService)

	// Middleware to inject logger with trace ID
	loggerMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = logger.EnrichContextWithLogger(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Wrap the router with OpenTelemetry middleware to create spans for each request
	handler := otelhttp.NewHandler(loggerMiddleware(router), "api")

	serverAddr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", cfg.ServerPort).Msg("API Service starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("listen")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}
