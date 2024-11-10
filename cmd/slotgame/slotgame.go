package main

import (
	"context"
	"fmt"
	"github.com/apolunin/slotgame/config"
	"github.com/apolunin/slotgame/internal/api"
	"github.com/apolunin/slotgame/internal/service"
	"github.com/apolunin/slotgame/internal/storage"
	"github.com/apolunin/slotgame/logger"
	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"

	_ "github.com/apolunin/slotgame/docs"
)

// @title           Slot Game API
// @version         1.0
// @description     Simple API simulating the behaviour of a slot machine.
// @termsOfService  http://swagger.io/terms/

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg := config.SlotGameConfig{}

	if err := env.Parse(&cfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to parse application config: %v", err)
		return
	}

	log, err := logger.NewLogger(cfg.LogCfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to configure application logger: %v", err)
		return
	}

	slog.SetDefault(log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("parsing config...")

	cfgConn, err := pgxpool.ParseConfig(cfg.DBCfg.URL())
	if err != nil {
		log.With(logger.FieldError, err).Error("failed to parse database configuration")
		return
	}

	log.Info("initializing database pool...")

	pool, err := pgxpool.NewWithConfig(ctx, cfgConn)
	if err != nil {
		slog.With(logger.FieldError, err).Error("failed to initialize database connection pool")
		return
	}

	log.Info("connecting to database...")

	if err = pool.Ping(ctx); err != nil {
		slog.With(logger.FieldError, err).Error("failed to connect to database")
		return
	}

	var (
		repo        = storage.NewStorage(pool)
		authService = service.NewAuthService([]byte(cfg.JWTSecret))
		userService = service.NewUserService(repo, authService)
		slotMachine = service.NewSlotMachine()
		slotService = service.NewSlotService(repo, slotMachine)
	)

	server := api.NewServer(authService, userService, slotService, cfg.HTTPCfg.RateLimitCfg)
	if err = server.Start(cfg.HTTPCfg.Port); err != nil {
		log.With(logger.FieldError, err).Error("failed to start server")
		return
	}
}
