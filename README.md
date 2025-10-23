# otelgen

A simple and reliable CLI tool for generating OpenTelemetry traces, metrics, and logs to test OTLP endpoints.

## Features

- Generate traces, metrics, and logs
- Support for gRPC, gRPCs, HTTP, and HTTPS protocols
- Automatic port selection (443 for gRPC/gRPCs/HTTPS, 80 for HTTP)
- Configurable rate and duration
- Simple and intuitive CLI interface

## Installation

### Build from source

```bash
go build -o otelgen ./cmd/otelgen
```

### Build Docker image

```bash
docker build -t otelgen .
```

The Docker image uses a multi-stage build with `scratch` as the final base image for minimal size (~20MB).

## Usage

### Traces

Generate trace data:

```bash
# Using gRPCs (secure gRPC)
./otelgen traces \
  --otlp-endpoint grpcs://example.com:443 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Using HTTP
./otelgen traces \
  --otlp-endpoint http://localhost:80 \
  --service my-service \
  --rate 10 \
  --duration 1m

# Port is optional - will use default 443 for gRPC
./otelgen traces \
  --otlp-endpoint grpc://localhost \
  --service my-service
```

### Metrics

Generate metrics data:

```bash
# Using gRPCs
./otelgen metrics \
  --otlp-endpoint grpcs://example.com:443 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Using HTTPS
./otelgen metrics \
  --otlp-endpoint https://localhost:443 \
  --service my-service \
  --rate 5 \
  --duration 30s
```

### Logs

Generate log data:

```bash
# Using gRPCs
./otelgen logs \
  --otlp-endpoint grpcs://example.com:443 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Using HTTP with custom port
./otelgen logs \
  --otlp-endpoint http://localhost:8080 \
  --service my-service \
  --rate 20 \
  --duration 5s
```

## Docker Usage

```bash
# Traces
docker run --rm otelgen traces \
  --otlp-endpoint grpcs://your-endpoint:443 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Metrics
docker run --rm otelgen metrics \
  --otlp-endpoint grpcs://your-endpoint:443 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Logs
docker run --rm otelgen logs \
  --otlp-endpoint grpcs://your-endpoint:443 \
  --service my-service \
  --rate 1 \
  --duration 10s
```

## Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `--otlp-endpoint` | OTLP endpoint URL (grpc://, grpcs://, http://, https://) | - | Yes |
| `--service` | Service name for telemetry | otelgen | No |
| `--rate` | Number of telemetry items per second | 1 | No |
| `--duration` | How long to generate telemetry (e.g., 10s, 1m, 1h) | 10s | No |
| `--size` | Payload size to increase data volume (e.g., 1kb, 1mb, 500b) | - | No |
| `--batch-size` | Maximum number of logs to batch before sending (logs only) | 512 | No |
| `--headers` | Additional headers (e.g., key1=value1,key2=value2) | - | No |
| `--verbose` | Enable verbose logging | false | No |
| `--insecure-skip-verify` | Skip TLS certificate verification (insecure) | false | No |

## Protocol Support

- `grpc://` - Insecure gRPC (default port: 443)
- `grpcs://` - Secure gRPC with TLS (default port: 443)
- `http://` - Insecure HTTP (default port: 80)
- `https://` - Secure HTTPS with TLS (default port: 443)

## Default Ports

If you don't specify a port in the endpoint URL, the following defaults are used:

- gRPC/gRPCs: 443
- HTTP: 80
- HTTPS: 443

## Examples

```bash
# Quick test with local collector
./otelgen traces --otlp-endpoint grpc://localhost --service test-app --duration 5s

# High-volume test
./otelgen metrics --otlp-endpoint http://localhost:80 --service load-test --rate 100 --duration 1m

# Production endpoint test
./otelgen logs --otlp-endpoint grpcs://prod.example.com:443 --service prod-app --rate 10 --duration 30s

# Test with increased payload size (1KB per trace)
./otelgen traces --otlp-endpoint grpc://localhost --service test-app --size 1kb --duration 10s

# Load test with large payloads (1MB per metric)
./otelgen metrics --otlp-endpoint http://localhost:80 --service load-test --size 1mb --rate 10 --duration 30s

# Test with custom payload size (500 bytes)
./otelgen logs --otlp-endpoint grpcs://example.com:443 --service test-app --size 500b --rate 5 --duration 1m

# Large logs with smaller batch size to avoid gRPC message size limit (4MB)
# For 1MB logs, use batch size of 3 to keep messages under 4MB
./otelgen logs --otlp-endpoint grpcs://example.com:443 --service test-app --size 1mb --rate 10 --batch-size 3 --duration 10s
```

## What Gets Generated

### Traces
- Parent spans with child spans
- Random operation types and IDs
- Realistic timing and nesting
- Optional payload padding via attributes when `--size` is specified

### Metrics
- Counter: `otelgen.requests`
- Histogram: `otelgen.duration`
- Gauge: `otelgen.cpu_usage`
- Optional payload padding via attributes when `--size` is specified

### Logs
- Proper OTLP log records with resource attributes
- Various log levels (INFO, WARN, ERROR, DEBUG) mapped to appropriate severity
- Log body contains realistic JSON structured data including:
  - Timestamp, service name, environment, version
  - HTTP request details (method, endpoint, status code, duration, user agent, client IP)
  - User information (ID, email, role, organization)
  - Error details with stack traces (for ERROR level)
  - Database query metrics (30% of logs)
- Additional attributes: component, request_id, user_id
- When `--size` is specified, the JSON body is expanded to reach target size
- **Batch Size**: Logs are batched before sending to improve efficiency. The default batch size is 512 logs. When using large log sizes (e.g., `--size=1mb`), you should reduce the batch size using `--batch-size` to avoid exceeding the gRPC message size limit (typically 4MB). For example, with 1MB logs, use `--batch-size=3` to keep messages under the limit.

#### Sample Log Output
```json
{
  "timestamp": "2025-01-23T10:15:30.123456789Z",
  "level": "INFO",
  "message": "Request processed successfully",
  "service": "api-gateway",
  "environment": "production",
  "version": "v1.2.3",
  "host": "server-3",
  "pod_id": "pod-42-aBcDeFgH",
  "request_id": "req-XyZ123AbC456-1737628530",
  "trace_id": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "span_id": "q1w2e3r4t5y6u7i8",
  "http": {
    "method": "POST",
    "endpoint": "/api/v1/orders",
    "status_code": 201,
    "duration_ms": 234,
    "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
    "client_ip": "10.142.15.89"
  },
  "user": {
    "id": "user_1234",
    "email": "user1234@example.com",
    "role": "user",
    "org_id": "org_56"
  },
  "database": {
    "query_time_ms": 45,
    "rows_affected": 1,
    "connection_id": 23,
    "database": "orders_db"
  }
}
```
