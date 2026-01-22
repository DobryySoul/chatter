package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func WithTracing(next http.Handler, tracer trace.Tracer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := r.Method + " " + normalizePath(r.URL.Path)
		ctx, span := tracer.Start(r.Context(), spanName,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", normalizePath(r.URL.Path)),
				attribute.String("http.target", r.URL.Path),
			),
		)

		recorder := newResponseRecorder(w)
		next.ServeHTTP(recorder, r.WithContext(ctx))

		span.SetAttributes(
			attribute.Int("http.status_code", recorder.status),
		)

		if recorder.status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(recorder.status))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	})
}
