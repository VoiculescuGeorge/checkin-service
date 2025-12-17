// Entry point for REST API
package main

import (
	"context"
	"log"
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
	"checkin.service/pkg/database"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
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

	// Initialize dependencies
	sqsClient := sqs.NewFromConfig(awsCfg)
	repo := repository.NewWorkingTimeRepository(db)
	producer := messaging.NewSQSProducer(sqsClient, cfg.LaborSQSQueueURL, cfg.EmailSQSQueueURL)
	coreService := checkin_service.NewCheckInService(repo, *producer)

	// Setup router and server
	router := api.NewRouter(*coreService)

	serverAddr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API Service starting on port %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
