package main

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/jackc/tern/v2/migrate"
	"github.com/skinkvi/effective_mobile/api/router"
	"github.com/skinkvi/effective_mobile/api/routes"
	"github.com/skinkvi/effective_mobile/internal/config"
	"github.com/skinkvi/effective_mobile/internal/logger"
	"github.com/skinkvi/effective_mobile/internal/storage/postgres"
	"go.uber.org/zap"

	_ "github.com/skinkvi/effective_mobile/docs"
)

// @title		Effective Mobile Sub Service API
// @version	1.0
// @host		localhost:8080
// @BasePath	/api

func main() {
	cfg := config.MustLoad("./config/local.yaml")

	log, err := logger.NewLogger(cfg)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx := context.Background()

	storage, err := postgres.New(ctx, cfg.DatabaseURL, log)
	if err != nil {
		log.Error("failed to init storage", zap.Error(err))
		os.Exit(1)
	}
	defer storage.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := storage.GetDB().Acquire(ctx)
		if err != nil {
			log.Fatal("failed to acquire connection from pool", zap.Error(err))
		}
		defer conn.Release()

		migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
		if err != nil {
			log.Fatal("failed to create migrator", zap.Error(err))
		}

		migrationsDir := "./migrations"
		if err := migrator.LoadMigrations(os.DirFS(migrationsDir)); err != nil {
			log.Fatal("failed to load migrations", zap.Error(err))
		}

		if err := migrator.Migrate(ctx); err != nil {
			log.Fatal("migration failed", zap.Error(err))
		}
		log.Info("migrations applied successfully")

	}()

	wg.Wait()

	log.Info("storage initialized")

	r := router.NewRouter(log)
	api := r.Group("/api")
	routes.SubscriptionRoutes(api, storage, log)

	if err := r.Run(); err != nil && err != http.ErrServerClosed {
		log.Fatal("failed to run server", zap.Error(err))
	}
}
