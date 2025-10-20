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
| `--insecure` | Use insecure connection (not implemented yet) | false | No |

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
```

## What Gets Generated

### Traces
- Parent spans with child spans
- Random operation types and IDs
- Realistic timing and nesting

### Metrics
- Counter: `otelgen.requests`
- Histogram: `otelgen.duration`
- Gauge: `otelgen.cpu_usage`

### Logs
- Various log levels (INFO, WARN, ERROR, DEBUG)
- Realistic log messages
- Additional attributes (component, request_id, user_id)
