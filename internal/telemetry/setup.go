// Package telemetry provides OpenTelemetry setup and instrumentation utilities.
//
// This package configures tracing, metrics, and logging using OpenTelemetry's
// auto-export functionality, which respects OTEL_* environment variables for
// configuration.
//
// Usage:
//
//	shutdown, err := telemetry.SetupOpenTelemetry(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(context.Background())
//
//	if err := telemetry.SetupSlog("daco-svc"); err != nil {
//	    log.Fatal(err)
//	}
//
// Configuration via environment variables:
//   - OTEL_TRACES_SAMPLER=traceidratio
//   - OTEL_TRACES_SAMPLER_ARG=0.1  (10% sampling)
//   - OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	setupOnce    sync.Once
	setupErr     error
	shutdownFunc func(context.Context) error
	setupMu      sync.RWMutex
)

// SetupOpenTelemetry initializes OpenTelemetry tracing, metrics, and logging.
// It configures exporters and sampling using the OTEL_* environment variables:
//
//   - OTEL_TRACES_EXPORTER: trace exporter (e.g., "otlp", "none")
//   - OTEL_METRICS_EXPORTER: metrics exporter (e.g., "otlp", "none")
//   - OTEL_LOGS_EXPORTER: log exporter (e.g., "otlp", "none")
//   - OTEL_TRACES_SAMPLER: sampler type (e.g., "traceidratio", "always_on", "always_off")
//   - OTEL_TRACES_SAMPLER_ARG: sampler argument (e.g., "0.1" for 10% sampling)
//   - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP endpoint URL
//
// The function returns a shutdown function that must be called to cleanly
// release resources, typically via defer. The shutdown function is safe to
// call multiple times and from multiple goroutines.
//
// This function can only be called once successfully. Subsequent calls will
// return the same shutdown function and error from the first call. It is safe
// to call this function concurrently from multiple goroutines.
//
// This function must be called before SetupSlog to ensure the log provider is
// initialized.
func SetupOpenTelemetry(ctx context.Context) (func(context.Context) error, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	setupOnce.Do(func() {
		var shutdownFuncs []func(context.Context) error
		var shutdownMu sync.Mutex
		var shutdownCalled bool

		shutdown := func(ctx context.Context) error {
			if ctx == nil {
				ctx = context.Background()
			}

			shutdownMu.Lock()
			defer shutdownMu.Unlock()

			if shutdownCalled {
				return nil
			}
			shutdownCalled = true

			var err error
			for _, fn := range shutdownFuncs {
				err = errors.Join(err, fn(ctx))
			}
			shutdownFuncs = nil
			return err
		}

		otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

		traceExporter, err := autoexport.NewSpanExporter(ctx)
		if err != nil {
			setupErr = errors.Join(
				fmt.Errorf("failed to initialize trace exporter: %w", err),
				shutdown(ctx))
			return
		}
		tp := trace.NewTracerProvider(
			trace.WithBatcher(traceExporter),
		)
		shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
		otel.SetTracerProvider(tp)

		metricReader, err := autoexport.NewMetricReader(ctx)
		if err != nil {
			setupErr = errors.Join(
				fmt.Errorf("failed to initialize metric reader: %w", err),
				shutdown(ctx))
			return
		}
		mp := metric.NewMeterProvider(
			metric.WithReader(metricReader),
		)
		shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
		otel.SetMeterProvider(mp)

		logExporter, err := autoexport.NewLogExporter(ctx)
		if err != nil {
			setupErr = errors.Join(
				fmt.Errorf("failed to initialize log exporter: %w", err),
				shutdown(ctx))
			return
		}
		lp := log.NewLoggerProvider(
			log.WithProcessor(log.NewBatchProcessor(logExporter)),
		)
		shutdownFuncs = append(shutdownFuncs, lp.Shutdown)
		global.SetLoggerProvider(lp)

		shutdownFunc = shutdown
	})

	return shutdownFunc, setupErr
}

// SetupSlog configures the default slog logger to use OpenTelemetry.
// It creates an OpenTelemetry-backed slog handler that bridges slog
// logs to OpenTelemetry's logging system.
//
// The serviceName parameter identifies the service in log records.
// This function should be called after SetupOpenTelemetry to ensure
// logs are properly exported. If called before SetupOpenTelemetry,
// logs will use a no-op provider and won't be exported. It is safe to
// call multiple times, but each call replaces the previous default
// logger.
func SetupSlog(serviceName string) {
	h := otelslog.NewHandler(
		serviceName,
		otelslog.WithLoggerProvider(global.GetLoggerProvider()),
		otelslog.WithSource(true),
	)
	slog.SetDefault(slog.New(h))
}
