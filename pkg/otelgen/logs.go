package otelgen

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
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

var endpoints = []string{
	"/api/v1/users",
	"/api/v1/orders",
	"/api/v1/products",
	"/api/v1/payments",
	"/api/v1/auth",
	"/api/v1/analytics",
	"/api/v1/notifications",
	"/api/v1/reports",
}

var httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
	"okhttp/4.9.0",
	"axios/0.21.1",
}

var errorMessages = []string{
	"connection timeout",
	"invalid authentication token",
	"rate limit exceeded",
	"database connection failed",
	"service unavailable",
	"invalid request payload",
	"resource not found",
	"permission denied",
}

// generateRealisticLogPayload creates a realistic JSON log payload
func generateRealisticLogPayload(baseMessage string, level string, targetSize int64) string {
	logData := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339Nano),
		"level":        level,
		"message":      baseMessage,
		"service":      "api-gateway",
		"environment":  "production",
		"version":      "v1.2.3",
		"host":         fmt.Sprintf("server-%d", rand.Intn(10)),
		"pod_id":       fmt.Sprintf("pod-%d-%s", rand.Intn(100), randomString(8)),
		"request_id":   fmt.Sprintf("req-%s-%d", randomString(16), time.Now().Unix()),
		"trace_id":     randomString(32),
		"span_id":      randomString(16),
		"http": map[string]interface{}{
			"method":      httpMethods[rand.Intn(len(httpMethods))],
			"endpoint":    endpoints[rand.Intn(len(endpoints))],
			"status_code": []int{200, 201, 400, 401, 403, 404, 500, 502, 503}[rand.Intn(9)],
			"duration_ms": rand.Intn(5000),
			"user_agent":  userAgents[rand.Intn(len(userAgents))],
			"client_ip":   fmt.Sprintf("10.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256)),
		},
		"user": map[string]interface{}{
			"id":       fmt.Sprintf("user_%d", rand.Intn(10000)),
			"email":    fmt.Sprintf("user%d@example.com", rand.Intn(10000)),
			"role":     []string{"admin", "user", "guest", "developer"}[rand.Intn(4)],
			"org_id":   fmt.Sprintf("org_%d", rand.Intn(100)),
		},
	}

	// Add error details for ERROR level
	if level == "ERROR" {
		logData["error"] = map[string]interface{}{
			"type":    "ServiceError",
			"message": errorMessages[rand.Intn(len(errorMessages))],
			"stack_trace": fmt.Sprintf("at com.example.service.Handler.handle(%s.java:%d)\n"+
				"at com.example.service.Processor.process(%s.java:%d)\n"+
				"at com.example.service.Worker.run(%s.java:%d)",
				randomString(10), rand.Intn(500),
				randomString(10), rand.Intn(500),
				randomString(10), rand.Intn(500)),
			"code": fmt.Sprintf("ERR_%d", rand.Intn(9999)),
		}
	}

	// Add database info occasionally
	if rand.Float32() < 0.3 {
		logData["database"] = map[string]interface{}{
			"query_time_ms": rand.Intn(1000),
			"rows_affected": rand.Intn(100),
			"connection_id": rand.Intn(50),
			"database":      []string{"users_db", "orders_db", "products_db", "analytics_db"}[rand.Intn(4)],
		}
	}

	// Marshal to JSON
	jsonBytes, _ := json.Marshal(logData)
	currentSize := int64(len(jsonBytes))

	// If we need to pad to reach target size, add a "payload_data" field
	if targetSize > 0 && currentSize < targetSize {
		remainingSize := targetSize - currentSize - 100 // Reserve space for field name and JSON structure
		if remainingSize > 0 {
			logData["payload_data"] = randomString(int(remainingSize))
			jsonBytes, _ = json.Marshal(logData)
		}
	}

	return string(jsonBytes)
}

// randomString generates a random alphanumeric string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// GenerateLogs generates log data and sends it to the specified OTLP endpoint
func GenerateLogs(endpoint *Endpoint, serviceName string, rate int, durationStr string, payloadSize int64, batchSize int, headers map[string]string, verbose bool, insecureSkip bool) error {
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

	// Create log exporter based on protocol
	var exporter sdklog.Exporter

	// Use context with timeout for exporter creation
	exporterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if endpoint.IsGRPC() {
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(endpoint.Address()),
		}

		if endpoint.Secure {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: insecureSkip,
				MinVersion:         tls.VersionTLS12,
			}
			if verbose {
				fmt.Printf("[VERBOSE] Using TLS with system certs, InsecureSkipVerify=%v\n", insecureSkip)
			}
			opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			if verbose {
				fmt.Println("[VERBOSE] Using insecure gRPC connection")
			}
			opts = append(opts, otlploggrpc.WithInsecure())
		}

		if len(headers) > 0 {
			if verbose {
				fmt.Printf("[VERBOSE] Adding headers: %v\n", headers)
			}
			opts = append(opts, otlploggrpc.WithHeaders(headers))
		}

		if verbose {
			fmt.Printf("[VERBOSE] Creating gRPC log exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlploggrpc.New(exporterCtx, opts...)
	} else {
		opts := []otlploghttp.Option{
			otlploghttp.WithEndpoint(endpoint.Address()),
		}

		if !endpoint.Secure {
			if verbose {
				fmt.Println("[VERBOSE] Using insecure HTTP connection")
			}
			opts = append(opts, otlploghttp.WithInsecure())
		} else {
			if verbose {
				fmt.Printf("[VERBOSE] Using HTTPS with system certs, InsecureSkipVerify=%v\n", insecureSkip)
			}
			if insecureSkip {
				opts = append(opts, otlploghttp.WithTLSClientConfig(&tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				}))
			}
		}

		if len(headers) > 0 {
			if verbose {
				fmt.Printf("[VERBOSE] Adding headers: %v\n", headers)
			}
			opts = append(opts, otlploghttp.WithHeaders(headers))
		}

		if verbose {
			fmt.Printf("[VERBOSE] Creating HTTP log exporter for %s\n", endpoint.Address())
		}
		exporter, err = otlploghttp.New(exporterCtx, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	if verbose {
		fmt.Println("[VERBOSE] Log exporter created successfully")
	}

	defer exporter.Shutdown(ctx)

	// Create batch processor with configurable batch size
	batchProcessor := sdklog.NewBatchProcessor(exporter,
		sdklog.WithMaxQueueSize(batchSize*2), // Queue size should be larger than batch size
		sdklog.WithExportMaxBatchSize(batchSize),
	)

	if verbose {
		fmt.Printf("[VERBOSE] Configured batch processor with max batch size: %d\n", batchSize)
	}

	// Create log provider
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(batchProcessor),
		sdklog.WithResource(res),
	)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := lp.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Error shutting down log provider: %v\n", err)
		}
	}()

	logger := lp.Logger("otelgen")

	// Generate logs
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	timer := time.NewTimer(duration)
	defer timer.Stop()

	count := 0
	for {
		select {
		case <-timer.C:
			fmt.Printf("Generated %d log records\n", count)
			return nil
		case <-ticker.C:
			generateLogRecord(ctx, logger, payloadSize)
			count++
		}
	}
}

func generateLogRecord(ctx context.Context, logger log.Logger, payloadSize int64) {
	baseMessage := logMessages[rand.Intn(len(logMessages))]
	level := logLevels[rand.Intn(len(logLevels))]

	// Generate realistic JSON log body
	var logBody string
	if payloadSize > 0 {
		logBody = generateRealisticLogPayload(baseMessage, level, payloadSize)
	} else {
		// For no size specified, still create a smaller realistic JSON log
		logBody = generateRealisticLogPayload(baseMessage, level, 0)
	}

	// Map log level to severity
	var severity log.Severity
	switch level {
	case "DEBUG":
		severity = log.SeverityDebug
	case "INFO":
		severity = log.SeverityInfo
	case "WARN":
		severity = log.SeverityWarn
	case "ERROR":
		severity = log.SeverityError
	default:
		severity = log.SeverityInfo
	}

	// Create attributes
	attrs := []log.KeyValue{
		log.String("component", "otelgen"),
		log.String("request_id", fmt.Sprintf("req-%s", randomString(16))),
		log.String("user_id", fmt.Sprintf("user_%d", rand.Intn(10000))),
	}

	// Emit log record with body as the message
	logRecord := log.Record{}
	logRecord.SetTimestamp(time.Now())
	logRecord.SetObservedTimestamp(time.Now())
	logRecord.SetSeverity(severity)
	logRecord.SetSeverityText(level)
	logRecord.SetBody(log.StringValue(logBody))
	logRecord.AddAttributes(attrs...)

	logger.Emit(ctx, logRecord)

	// Also print to stdout
	slog.Info("Generated log", "level", level, "message", baseMessage)
}
