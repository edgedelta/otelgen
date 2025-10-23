package otelgen

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// GenerateTraces generates trace data and sends it to the specified OTLP endpoint
func GenerateTraces(endpoint *Endpoint, serviceName string, rate int, durationStr string, payloadSize int64, headers map[string]string, verbose bool, insecureSkip bool) error {
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

	// Test network connectivity first
	if verbose {
		fmt.Printf("[VERBOSE] Testing network connectivity to %s...\n", endpoint.Address())
		testCtx, testCancel := context.WithTimeout(ctx, 5*time.Second)
		defer testCancel()

		dialer := &net.Dialer{}
		conn, err := dialer.DialContext(testCtx, "tcp", endpoint.Address())
		if err != nil {
			fmt.Printf("[VERBOSE] WARNING: Cannot establish TCP connection: %v\n", err)
		} else {
			fmt.Printf("[VERBOSE] TCP connection successful\n")
			conn.Close()
		}
	}

	// Create exporter based on protocol
	var exporter sdktrace.SpanExporter

	// Use context with timeout for exporter creation
	exporterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if endpoint.IsGRPC() {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint.Address()),
		}

		if endpoint.Secure {
			// Create TLS config with system cert pool
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

		// Add gRPC dial options for better debugging and connection management
		dialOpts := []grpc.DialOption{
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                10 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		}

		if verbose {
			fmt.Printf("[VERBOSE] Adding gRPC keepalive and timeout options\n")
		}

		opts = append(opts, otlptracegrpc.WithDialOption(dialOpts...))

		if verbose {
			fmt.Printf("[VERBOSE] Creating gRPC trace exporter for %s\n", endpoint.Address())
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
			fmt.Printf("[VERBOSE] Creating HTTP trace exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlptracehttp.New(exporterCtx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	if verbose {
		fmt.Println("[VERBOSE] Trace exporter created successfully")
		fmt.Println("[VERBOSE] Attempting to export a test span to verify connectivity...")

		// Test export with a single span using a separate test provider
		testTP := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter,
				sdktrace.WithBatchTimeout(1*time.Second),
				sdktrace.WithExportTimeout(10*time.Second),
			),
			sdktrace.WithResource(res),
		)
		testTracer := testTP.Tracer("test")

		_, testSpan := testTracer.Start(ctx, "connection-test")
		testSpan.End()

		// Force flush to send immediately
		flushCtx, flushCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer flushCancel()

		fmt.Println("[VERBOSE] Flushing test span...")
		if err := testTP.ForceFlush(flushCtx); err != nil {
			fmt.Printf("[VERBOSE] WARNING: Test span export failed: %v\n", err)
			fmt.Printf("[VERBOSE] This indicates connectivity or authentication issues with the endpoint\n")
		} else {
			fmt.Printf("[VERBOSE] Test span exported successfully! Endpoint is reachable and accepting data.\n")
		}

		// Shutdown test provider but keep exporter alive
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		testTP.Shutdown(shutdownCtx)

		// Create a new exporter for actual trace generation since we used this one for testing
		fmt.Println("[VERBOSE] Creating new exporter for trace generation...")
		exporterCtx2, cancel2 := context.WithTimeout(ctx, 10*time.Second)
		defer cancel2()

		if endpoint.IsGRPC() {
			opts := []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint(endpoint.Address()),
			}

			if endpoint.Secure {
				tlsConfig := &tls.Config{
					InsecureSkipVerify: insecureSkip,
					MinVersion:         tls.VersionTLS12,
				}
				opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
			} else {
				opts = append(opts, otlptracegrpc.WithInsecure())
			}

			if len(headers) > 0 {
				opts = append(opts, otlptracegrpc.WithHeaders(headers))
			}

			dialOpts := []grpc.DialOption{
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                10 * time.Second,
					Timeout:             5 * time.Second,
					PermitWithoutStream: true,
				}),
			}
			opts = append(opts, otlptracegrpc.WithDialOption(dialOpts...))

			exporter, err = otlptracegrpc.New(exporterCtx2, opts...)
		} else {
			opts := []otlptracehttp.Option{
				otlptracehttp.WithEndpoint(endpoint.Address()),
			}

			if !endpoint.Secure {
				opts = append(opts, otlptracehttp.WithInsecure())
			} else if insecureSkip {
				opts = append(opts, otlptracehttp.WithTLSClientConfig(&tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				}))
			}

			if len(headers) > 0 {
				opts = append(opts, otlptracehttp.WithHeaders(headers))
			}

			exporter, err = otlptracehttp.New(exporterCtx2, opts...)
		}

		if err != nil {
			return fmt.Errorf("failed to create new trace exporter: %w", err)
		}

		fmt.Println()
	}

	defer exporter.Shutdown(ctx)

	// Create trace provider with configurable timeouts
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(2*time.Second),
			sdktrace.WithExportTimeout(30*time.Second), // Increased timeout for slow connections
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
	)
	defer func() {
		// Give it time to flush remaining spans
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Error shutting down trace provider: %v\n", err)
		}
	}()

	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("otelgen")

	// Generate traces
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	timer := time.NewTimer(duration)
	defer timer.Stop()

	count := 0
	for {
		select {
		case <-timer.C:
			fmt.Printf("Generated %d traces\n", count)
			return nil
		case <-ticker.C:
			if err := generateTrace(ctx, tracer, payloadSize); err != nil {
				fmt.Printf("Error generating trace: %v\n", err)
			}
			count++
		}
	}
}

func generateTrace(ctx context.Context, tracer trace.Tracer, payloadSize int64) error {
	// Create attributes list
	attrs := []attribute.KeyValue{
		attribute.String("operation.type", "http"),
		attribute.Int("operation.id", rand.Intn(1000)),
	}

	// Add padding attribute if size is specified
	if payloadSize > 0 {
		attrs = append(attrs, attribute.String("payload.data", GeneratePadding(payloadSize)))
	}

	// Create a parent span
	ctx, span := tracer.Start(ctx, "parent-operation",
		trace.WithAttributes(attrs...))
	defer span.End()

	// Simulate some work
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))

	// Create child spans
	for i := 0; i < rand.Intn(3)+1; i++ {
		childAttrs := []attribute.KeyValue{
			attribute.String("child.type", "db"),
			attribute.Int("child.id", i),
		}

		// Add padding to child spans as well if size is specified
		if payloadSize > 0 {
			childAttrs = append(childAttrs, attribute.String("payload.data", GeneratePadding(payloadSize)))
		}

		_, childSpan := tracer.Start(ctx, fmt.Sprintf("child-operation-%d", i),
			trace.WithAttributes(childAttrs...))
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)))
		childSpan.End()
	}

	return nil
}
