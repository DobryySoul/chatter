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
	"context"
	"net/http"

	"go.uber.org/zap"
)

func Run(cfg *config.Config, logger *zap.Logger) {
	pgpool, err := postgres.New(context.Background(), &cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to postgres", zap.Error(err))
	}

	logger.Info("Connected to postgres")

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

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)

	mux.Handle("POST /rooms", middleware.RequireAuth(authManager, http.HandlerFunc(signalingHandler.CreateRoom)))
	mux.HandleFunc("GET /ws/", signalingHandler.JoinRoom)

	server := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: withCORS(mux),
	}

	logger.Info("Signaling server started", zap.String("addr", cfg.Server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Signaling server failed", zap.Error(err))
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
