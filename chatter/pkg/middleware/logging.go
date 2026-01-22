package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func WithRequestLogging(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)

			return
		}

		start := time.Now()
		recorder := newResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", recorder.status),
			zap.Int("bytes", recorder.bytes),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote", r.RemoteAddr),
		}

		if userID, ok := UserIDFromContext(r.Context()); ok && userID != 0 {
			fields = append(fields, zap.Uint64("userID", userID))
		}

		if username, ok := UserFromContext(r.Context()); ok && username != "" {
			fields = append(fields, zap.String("username", username))
		}

		logger.Info("HTTP request", fields...)
	})
}
