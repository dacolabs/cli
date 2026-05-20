package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func Telemetry() func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware(
		"", // overridden per-request by SpanNameFormatter below.
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		otelhttp.WithMeterProvider(otel.GetMeterProvider()),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			if r.Pattern != "" {
				return r.Pattern
			}
			return r.Method + " " + r.URL.Path
		}),
		otelhttp.WithSpanOptions(trace.WithSpanKind(trace.SpanKindServer)),
	)
}
