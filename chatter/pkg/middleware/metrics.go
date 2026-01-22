package middleware

import (
	"expvar"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	httpRequestsTotal     = expvar.NewMap("http_requests_total")
	httpRequestDurationMs = expvar.NewMap("http_request_duration_ms_total")
	httpResponseSizeBytes = expvar.NewMap("http_response_size_bytes_total")
)

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := newResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(recorder.status)

		labelKey := formatLabels(r.Method, path, status)
		httpRequestsTotal.Add(labelKey, 1)
		httpRequestDurationMs.Add(labelKey, time.Since(start).Milliseconds())
		httpResponseSizeBytes.Add(labelKey, int64(recorder.bytes))
	})
}

func normalizePath(path string) string {
	if strings.HasPrefix(path, "/ws/") && path != "/ws/" {
		return "/ws/:room"
	}

	return path
}

func formatLabels(method, path, status string) string {
	return method + "|" + path + "|" + status
}
