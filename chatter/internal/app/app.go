package app

import (
	"chatter/internal/config"
	"chatter/internal/handler"
	"chatter/internal/infra"
	"chatter/internal/repository"
	"chatter/internal/signaling"
	"chatter/internal/usecase"
	"chatter/pkg/middleware"
	"chatter/pkg/postgres"
	"chatter/pkg/redis"
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	pgpool, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to postgres", zap.Error(err))
	}

	logger.Info("Connected to postgres")

	rdb, err := redis.NewClient(ctx, &cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to redis", zap.Error(err))
	}

	// заглушка, redis пока что не используется
	_ = rdb

	logger.Info("Connected to redis")

	authStore := repository.NewAuthRepository(pgpool, logger)
	tokenStore := repository.NewRefreshTokenRepository(pgpool, logger)
	authManager := infra.NewJWTManager(cfg.Auth.Secret, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	authService := usecase.NewAuthService(authStore, tokenStore, authManager, cfg.Auth.RefreshTTL, logger)
	authHandler := handler.NewHandler(authService, logger)

	registry := signaling.NewRegistry(logger)
	signalingHandler := signaling.NewHandler(registry, authManager, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.Handle("GET /auth/sessions", middleware.RequireAuth(authManager, http.HandlerFunc(authHandler.Sessions)))

	mux.Handle("POST /rooms", middleware.RequireAuth(authManager, http.HandlerFunc(signalingHandler.CreateRoom)))
	mux.HandleFunc("GET /ws/", signalingHandler.JoinRoom)

	server := &http.Server{
		Addr: cfg.Server.Addr,
		Handler: middleware.WithTracing(
			middleware.WithRequestLogging(
				middleware.WithMetrics(middleware.WithCORS(mux, cfg.Server.CorsOrigins)),
				logger,
			),
			otel.Tracer("chatter/http"),
		),
	}

	logger.Info("Signaling server started", zap.String("addr", cfg.Server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Signaling server failed", zap.Error(err))
	}
}
