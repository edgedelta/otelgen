package otelgen

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
)

// GenerateMetrics generates metric data and sends it to the specified OTLP endpoint
func GenerateMetrics(endpoint *Endpoint, serviceName string, rate int, durationStr string, headers map[string]string, verbose bool, insecureSkip bool) error {
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

	// Create exporter based on protocol
	var exporter sdkmetric.Exporter

	// Use context with timeout for exporter creation
	exporterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if endpoint.IsGRPC() {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(endpoint.Address()),
		}

		if endpoint.Secure {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: insecureSkip,
				MinVersion:         tls.VersionTLS12,
			}
			if verbose {
				fmt.Printf("[VERBOSE] Using TLS with system certs, InsecureSkipVerify=%v\n", insecureSkip)
			}
			opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			if verbose {
				fmt.Println("[VERBOSE] Using insecure gRPC connection")
			}
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}

		if len(headers) > 0 {
			if verbose {
				fmt.Printf("[VERBOSE] Adding headers: %v\n", headers)
			}
			opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
		}

		if verbose {
			fmt.Printf("[VERBOSE] Creating gRPC metrics exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlpmetricgrpc.New(exporterCtx, opts...)
	} else {
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(endpoint.Address()),
		}

		if !endpoint.Secure {
			if verbose {
				fmt.Println("[VERBOSE] Using insecure HTTP connection")
			}
			opts = append(opts, otlpmetrichttp.WithInsecure())
		} else {
			if verbose {
				fmt.Printf("[VERBOSE] Using HTTPS with system certs, InsecureSkipVerify=%v\n", insecureSkip)
			}
			if insecureSkip {
				opts = append(opts, otlpmetrichttp.WithTLSClientConfig(&tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				}))
			}
		}

		if len(headers) > 0 {
			if verbose {
				fmt.Printf("[VERBOSE] Adding headers: %v\n", headers)
			}
			opts = append(opts, otlpmetrichttp.WithHeaders(headers))
		}

		if verbose {
			fmt.Printf("[VERBOSE] Creating HTTP metrics exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlpmetrichttp.New(exporterCtx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	if verbose {
		fmt.Println("[VERBOSE] Metrics exporter created successfully")
		fmt.Println("[VERBOSE] Note: Metrics will be exported periodically every 2 seconds")
		fmt.Println()
	}

	// Create meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(2*time.Second),
			sdkmetric.WithTimeout(30*time.Second), // Increased timeout
		)),
		sdkmetric.WithResource(res),
	)
	defer func() {
		if verbose {
			fmt.Println("[VERBOSE] Shutting down meter provider and flushing metrics...")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := mp.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Error shutting down meter provider: %v\n", err)
		} else if verbose {
			fmt.Println("[VERBOSE] Meter provider shut down successfully")
		}
	}()

	otel.SetMeterProvider(mp)
	meter := mp.Meter("otelgen")

	// Create metrics
	counter, err := meter.Int64Counter(
		"otelgen.requests",
		metric.WithDescription("Number of requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create counter: %w", err)
	}

	histogram, err := meter.Float64Histogram(
		"otelgen.duration",
		metric.WithDescription("Request duration"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return fmt.Errorf("failed to create histogram: %w", err)
	}

	gauge, err := meter.Float64ObservableGauge(
		"otelgen.cpu_usage",
		metric.WithDescription("CPU usage percentage"),
		metric.WithFloat64Callback(func(ctx context.Context, observer metric.Float64Observer) error {
			observer.Observe(rand.Float64()*100, metric.WithAttributes(
				attribute.String("host", "localhost"),
			))
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gauge: %w", err)
	}
	_ = gauge // gauge is automatically recorded

	// Generate metrics
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	timer := time.NewTimer(duration)
	defer timer.Stop()

	count := 0
	for {
		select {
		case <-timer.C:
			fmt.Printf("Generated %d metric events\n", count)

			// Force flush before returning to ensure all metrics are sent
			if verbose {
				fmt.Println("[VERBOSE] Forcing final metrics flush...")
			}
			flushCtx, flushCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer flushCancel()

			if err := mp.ForceFlush(flushCtx); err != nil {
				fmt.Printf("Warning: Failed to flush final metrics: %v\n", err)
			} else if verbose {
				fmt.Println("[VERBOSE] Final metrics flushed successfully")
			}

			return nil
		case <-ticker.C:
			// Record counter
			counter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("method", "GET"),
				attribute.String("endpoint", "/api/test"),
			))

			// Record histogram
			histogram.Record(ctx, rand.Float64()*1000, metric.WithAttributes(
				attribute.String("method", "GET"),
				attribute.String("endpoint", "/api/test"),
			))

			count++

			if verbose && count%5 == 0 {
				fmt.Printf("[VERBOSE] Generated %d metric events (next export in ~%ds)\n", count, 2-(count%2))
			}
		}
	}
}
