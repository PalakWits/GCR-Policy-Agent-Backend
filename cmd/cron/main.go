package main

import (
	"context"
	"flag"

	"adapter/internal/config"
	sellerDomain "adapter/internal/domain/seller"
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/database"
	"adapter/internal/shared/log"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// Define a command-line flag to trigger the job immediately
	runNow := flag.Bool("run-now", false, "Run the job once immediately and exit")
	flag.Parse()
	ctx := context.Background()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal(ctx, err, "Error loading .env file")
	}

	// Initialize logger
	log.InitLogger(log.Config{Level: "info", Destinations: []log.Destination{{Type: "stdout"}}})

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(ctx, err, "Error loading configuration")
	}

	// Initialize database
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(ctx, err, "Failed to connect to database")
	}

	// Create repository and service
	sellerRepo := sellerPorts.NewSellerRepository(db)
	sellerService := sellerDomain.NewSellerService(sellerRepo, cfg)

	// If the -run-now flag is provided, run the job once and exit
	if *runNow {
		log.Info(ctx, "Starting ONDC seller lookup job manually...")
		_, err := sellerService.SyncRegistry(sellerPorts.SellerRegistrySyncRequest{
			Domains: cfg.Domains,
		})
		if err != nil {
			log.Error(ctx, err, "ONDC seller lookup cron job failed")
		} else {
			log.Info(ctx, "ONDC seller lookup cron job completed successfully.")
		}
		return
	}

	// --- Original cron job scheduling ---
	// Initialize cron job
	c := cron.New()

	// Schedule the job to run every 6 hours
	c.AddFunc("@every 6h", func() {
		log.Info(ctx, "Starting ONDC seller lookup cron job...")
		_, err := sellerService.SyncRegistry(sellerPorts.SellerRegistrySyncRequest{
			Domains: cfg.Domains,
		})
		if err != nil {
			log.Error(ctx, err, "ONDC seller lookup cron job failed")
		} else {
			log.Info(ctx, "ONDC seller lookup cron job completed successfully.")
		}

	})

	log.Info(ctx, "Starting cron scheduler...")
	c.Start()

	// Keep the application running
	select {}
}
