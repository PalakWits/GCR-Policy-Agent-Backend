package di

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"adapter/internal/config"
	broadcastDomain "adapter/internal/domain/broadcast"
	buyerDomain "adapter/internal/domain/buyer"
	sellerDomain "adapter/internal/domain/seller"
	broadcastHandler "adapter/internal/handlers/broadcast"
	buyerHandler "adapter/internal/handlers/buyer"
	sellerHandler "adapter/internal/handlers/seller"
	buyerPorts "adapter/internal/ports/buyer"
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/caching"
	db "adapter/internal/shared/database"
	logger "adapter/internal/shared/log"
	// redisClient "adapter/internal/shared/redis"
)

type Container struct {
	Config           *config.Config
	DB               *gorm.DB
	CacheService     caching.CacheService
	SellerHandler    *sellerHandler.SellerHandler
	BuyerHandler     *buyerHandler.BuyerHandler
	BroadcastHandler *broadcastHandler.BroadcastHandler
}

func (c *Container) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Shutting down container resources...")

	if c.DB != nil {
		if err := db.Close(); err != nil {
			logger.Error(ctx, err, "Failed to close database connection")
		}
	}

	logger.Info(ctx, "Container shutdown complete")
	return nil
}

func InitContainer() (*Container, error) {
	cfg, err := config.LoadConfig()
	ctx := context.Background()
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("failed to load config: %w", err), "Configuration error")
	}

	database, err := db.Init(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("failed to initialize database: %w", err), "Database initialization error")
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info(ctx, "Running database migrations...")
	if err := database.AutoMigrate(&sellerPorts.Seller{}, &buyerPorts.Bap{}, &sellerPorts.SellerCatalogState{}, &buyerPorts.BapAccessPolicy{}, &buyerPorts.PermissionsJob{}); err != nil {
		logger.Fatal(ctx, err, "Failed to run database migrations")
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}
	logger.Info(ctx, "Database migrations completed successfully")

	sellerRepo := sellerPorts.NewSellerRepository(database)
	sellerService := sellerDomain.NewSellerService(sellerRepo, cfg)
	sellerHandler := sellerHandler.NewSellerHandler(sellerService)

	buyerRepo := buyerPorts.NewBuyerRepository(database)
	buyerService := buyerDomain.NewBuyerService(buyerRepo)
	buyerHandler := buyerHandler.NewBuyerHandler(buyerService)

	broadcastService := broadcastDomain.NewBroadcastService(buyerRepo, sellerRepo, cfg)
	broadcastHandler := broadcastHandler.NewBroadcastHandler(broadcastService)

	return &Container{
		Config:           cfg,
		DB:               database,
		SellerHandler:    sellerHandler,
		BuyerHandler:     buyerHandler,
		BroadcastHandler: broadcastHandler,
	}, err
}
