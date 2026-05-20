package telemetry

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestSetupOpenTelemetry(t *testing.T) {
	resetForTesting()
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx)
	if err != nil {
		t.Fatalf("SetupOpenTelemetry() failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("SetupOpenTelemetry() returned nil shutdown function")
	}

	// Verify shutdown doesn't panic (may fail to upload if no collector running)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown may error if no OTEL collector is running, which is expected in tests
	_ = shutdown(shutdownCtx)
}

func TestSetupOpenTelemetry_ShutdownIdempotent(t *testing.T) {
	resetForTesting()
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx)
	if err != nil {
		t.Fatalf("SetupOpenTelemetry() failed: %v", err)
	}

	// Call shutdown multiple times - should be safe even if it errors
	for range 3 {
		// May error without collector, but should not panic
		_ = shutdown(context.Background())
	}
}

func TestSetupSlog(t *testing.T) {
	resetForTesting()
	// Capture the original default logger
	original := slog.Default()
	defer slog.SetDefault(original)

	// Need to setup OpenTelemetry first
	shutdown, err := SetupOpenTelemetry(context.Background())
	if err != nil {
		t.Fatalf("SetupOpenTelemetry() failed: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	SetupSlog("test-service")

	// Verify a new logger was set
	logger := slog.Default()
	if logger == nil {
		t.Fatal("SetupSlog() did not set a default logger")
	}

	// Verify we can log without panicking
	logger.Info("test log message", "service", "test-service")
}

func TestSetupSlog_EmptyServiceName(t *testing.T) {
	resetForTesting()
	// Capture the original default logger
	original := slog.Default()
	defer slog.SetDefault(original)

	// Need to setup OpenTelemetry first
	shutdown, err := SetupOpenTelemetry(context.Background())
	if err != nil {
		t.Fatalf("SetupOpenTelemetry() failed: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	// Should not panic with empty service name
	SetupSlog("")

	logger := slog.Default()
	if logger == nil {
		t.Fatal("SetupSlog() with empty name did not set a default logger")
	}
}

func TestSetupSlog_WithoutOpenTelemetry(t *testing.T) {
	// Capture the original default logger
	original := slog.Default()
	defer slog.SetDefault(original)

	// SetupSlog should still work even without explicit OpenTelemetry
	// setup because global.GetLoggerProvider() returns a no-op provider
	// by default.
	SetupSlog("test-service")

	logger := slog.Default()
	if logger == nil {
		t.Fatal("SetupSlog() did not set a default logger")
	}
}

func TestSetupOpenTelemetry_NilContext(t *testing.T) {
	resetForTesting()
	// Intentionally exercising the nil-context error path. Using a
	// typed-nil variable avoids tripping SA1012 on a literal `nil`
	// argument while preserving the test intent.
	var nilCtx context.Context
	_, err := SetupOpenTelemetry(nilCtx)
	if err == nil {
		t.Fatal("SetupOpenTelemetry() with nil context should return error")
	}

	expectedMsg := "context cannot be nil"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestSetupOpenTelemetry_CalledTwice(t *testing.T) {
	resetForTesting()
	ctx := context.Background()

	// First call
	shutdown1, err1 := SetupOpenTelemetry(ctx)
	if err1 != nil {
		t.Fatalf("first SetupOpenTelemetry() failed: %v", err1)
	}
	if shutdown1 == nil {
		t.Fatal("first SetupOpenTelemetry() returned nil shutdown")
	}

	// Second call should return same shutdown function
	shutdown2, err2 := SetupOpenTelemetry(ctx)
	if err2 != err1 {
		t.Errorf("second call returned different error: %v vs %v", err2, err1)
	}
	if shutdown2 == nil {
		t.Fatal("second SetupOpenTelemetry() returned nil shutdown")
	}

	// Cleanup
	_ = shutdown1(context.Background())
}

func TestSetupOpenTelemetry_ConcurrentCalls(t *testing.T) {
	resetForTesting()
	const goroutines = 10
	var wg sync.WaitGroup
	results := make([]func(context.Context) error, goroutines)
	errors := make([]error, goroutines)

	ctx := context.Background()

	for i := range goroutines {
		wg.Go(func() {
			results[i], errors[i] = SetupOpenTelemetry(ctx)
		})
	}

	wg.Wait()

	// All calls should return the same shutdown function and error
	firstShutdown := results[0]
	firstErr := errors[0]

	for i := 1; i < goroutines; i++ {
		if errors[i] != firstErr {
			t.Errorf("goroutine %d got different error: %v vs %v", i, errors[i], firstErr)
		}
		// Note: Can't directly compare functions, but they should all be non-nil or all nil
		if (results[i] == nil) != (firstShutdown == nil) {
			t.Errorf("goroutine %d got different shutdown nil status", i)
		}
	}

	// Cleanup
	if firstShutdown != nil {
		_ = firstShutdown(context.Background())
	}
}

func TestShutdown_ConcurrentCalls(t *testing.T) {
	resetForTesting()
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx)
	if err != nil {
		t.Fatalf("SetupOpenTelemetry() failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("SetupOpenTelemetry() returned nil shutdown")
	}

	const goroutines = 10
	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			// Should be safe to call concurrently
			_ = shutdown(context.Background())
		})
	}

	wg.Wait()
}

func TestShutdown_WithNilContext(t *testing.T) {
	resetForTesting()
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx)
	if err != nil {
		t.Fatalf("SetupOpenTelemetry() failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("SetupOpenTelemetry() returned nil shutdown")
	}

	// Should not panic with nil context
	err = shutdown(nil)
	// May error without collector, but should not panic
	_ = err
}

// TestSetupOpenTelemetry_*ExporterError drives the autoexport error
// branches inside setupOnce.Do by pointing each exporter env var at a
// value autoexport doesn't recognize. Each branch short-circuits the
// rest of setup; setupOnce caches the error so subsequent calls
// return identically. resetForTesting clears the once between tests.
//
// The three tests are near-duplicates because the three error
// branches in SetupOpenTelemetry are themselves near-duplicates.
func TestSetupOpenTelemetry_TracesExporterError(t *testing.T) {
	resetForTesting()
	t.Setenv("OTEL_TRACES_EXPORTER", "definitely-not-a-real-exporter")
	t.Cleanup(resetForTesting)

	_, err := SetupOpenTelemetry(context.Background())
	if err == nil {
		t.Fatal("SetupOpenTelemetry: want error from invalid OTEL_TRACES_EXPORTER, got nil")
	}
}

func TestSetupOpenTelemetry_MetricsExporterError(t *testing.T) {
	resetForTesting()
	t.Setenv("OTEL_METRICS_EXPORTER", "definitely-not-a-real-exporter")
	t.Cleanup(resetForTesting)

	_, err := SetupOpenTelemetry(context.Background())
	if err == nil {
		t.Fatal("SetupOpenTelemetry: want error from invalid OTEL_METRICS_EXPORTER, got nil")
	}
}

func TestSetupOpenTelemetry_LogsExporterError(t *testing.T) {
	resetForTesting()
	t.Setenv("OTEL_LOGS_EXPORTER", "definitely-not-a-real-exporter")
	t.Cleanup(resetForTesting)

	_, err := SetupOpenTelemetry(context.Background())
	if err == nil {
		t.Fatal("SetupOpenTelemetry: want error from invalid OTEL_LOGS_EXPORTER, got nil")
	}
}

// resetForTesting resets the package state for testing purposes.
// This should only be called from tests.
func resetForTesting() {
	setupMu.Lock()
	defer setupMu.Unlock()

	setupOnce = sync.Once{}
	setupErr = nil
	shutdownFunc = nil
}
