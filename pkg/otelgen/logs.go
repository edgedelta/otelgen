package otelgen

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

var logMessages = []string{
	"Request processed successfully",
	"Database query executed",
	"Cache hit for key",
	"User authentication completed",
	"API request received",
	"Processing payment transaction",
	"Sending notification",
	"Background job started",
	"File uploaded successfully",
	"Configuration reloaded",
}

var logLevels = []string{
	"INFO",
	"WARN",
	"ERROR",
	"DEBUG",
}

// GenerateLogs generates log data and sends it to the specified OTLP endpoint
func GenerateLogs(endpoint *Endpoint, serviceName string, rate int, durationStr string, headers map[string]string, verbose bool, insecureSkip bool) error {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	ctx := context.Background()

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter to send logs as span events
	// This is a workaround since OTLP log exporters are still experimental
	var exporter sdktrace.SpanExporter

	// Use context with timeout for exporter creation
	exporterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if endpoint.IsGRPC() {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint.Address()),
		}

		if endpoint.Secure {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: insecureSkip,
				MinVersion:         tls.VersionTLS12,
			}
			if verbose {
				fmt.Printf("[VERBOSE] Using TLS with system certs, InsecureSkipVerify=%v\n", insecureSkip)
			}
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			if verbose {
				fmt.Println("[VERBOSE] Using insecure gRPC connection")
			}
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		if len(headers) > 0 {
			if verbose {
				fmt.Printf("[VERBOSE] Adding headers: %v\n", headers)
			}
			opts = append(opts, otlptracegrpc.WithHeaders(headers))
		}

		if verbose {
			fmt.Printf("[VERBOSE] Creating gRPC log exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlptracegrpc.New(exporterCtx, opts...)
	} else {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint.Address()),
		}

		if !endpoint.Secure {
			if verbose {
				fmt.Println("[VERBOSE] Using insecure HTTP connection")
			}
			opts = append(opts, otlptracehttp.WithInsecure())
		} else {
			if verbose {
				fmt.Printf("[VERBOSE] Using HTTPS with system certs, InsecureSkipVerify=%v\n", insecureSkip)
			}
			if insecureSkip {
				opts = append(opts, otlptracehttp.WithTLSClientConfig(&tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				}))
			}
		}

		if len(headers) > 0 {
			if verbose {
				fmt.Printf("[VERBOSE] Adding headers: %v\n", headers)
			}
			opts = append(opts, otlptracehttp.WithHeaders(headers))
		}

		if verbose {
			fmt.Printf("[VERBOSE] Creating HTTP log exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlptracehttp.New(exporterCtx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	if verbose {
		fmt.Println("[VERBOSE] Log exporter created successfully")
	}

	defer exporter.Shutdown(ctx)

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(2*time.Second),
			sdktrace.WithExportTimeout(30*time.Second), // Increased timeout
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
	)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Error shutting down trace provider: %v\n", err)
		}
	}()

	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("otelgen")

	// Generate logs as span events
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	timer := time.NewTimer(duration)
	defer timer.Stop()

	count := 0
	for {
		select {
		case <-timer.C:
			fmt.Printf("Generated %d log events\n", count)
			return nil
		case <-ticker.C:
			generateLogEvent(ctx, tracer)
			count++
		}
	}
}

func generateLogEvent(ctx context.Context, tracer trace.Tracer) {
	message := logMessages[rand.Intn(len(logMessages))]
	level := logLevels[rand.Intn(len(logLevels))]

	// Create a span to represent the log entry
	_, span := tracer.Start(ctx, "log."+level,
		trace.WithAttributes(
			attribute.String("log.message", message),
			attribute.String("log.level", level),
			attribute.String("component", "otelgen"),
			attribute.Int("request_id", rand.Intn(10000)),
			attribute.String("user_id", fmt.Sprintf("user_%d", rand.Intn(100))),
		))

	// Add event to represent the actual log
	span.AddEvent(message, trace.WithAttributes(
		attribute.String("severity", level),
	))

	span.End()

	// Also print to stdout
	slog.Info("Generated log", "level", level, "message", message)
}
