# Telemetrygen

A simple and reliable CLI tool for generating OpenTelemetry traces, metrics, and logs to test OTLP endpoints.

## Features

- Generate traces, metrics, and logs
- Support for gRPC, gRPCs, HTTP, and HTTPS protocols
- Automatic port selection (4317 for gRPC, 4318 for HTTP)
- Configurable rate and duration
- Simple and intuitive CLI interface

## Installation

### Build from source

```bash
go build -o telemetrygen
```

### Build Docker image

```bash
docker build -t telemetrygen .
```

The Docker image uses a multi-stage build with `scratch` as the final base image for minimal size (~20MB).

## Usage

### Traces

Generate trace data:

```bash
# Using gRPCs (secure gRPC)
./telemetrygen traces \
  --otlp-endpoint grpcs://852e95f4-5bbc-4d54-9879-b8591ff6194c-grpc-us-west2-cf.aws-staging.edgedelta.com:4317 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Using HTTP
./telemetrygen traces \
  --otlp-endpoint http://localhost:4318 \
  --service my-service \
  --rate 10 \
  --duration 1m

# Port is optional - will use default 4317 for gRPC
./telemetrygen traces \
  --otlp-endpoint grpc://localhost \
  --service my-service
```

### Metrics

Generate metrics data:

```bash
# Using gRPCs
./telemetrygen metrics \
  --otlp-endpoint grpcs://852e95f4-5bbc-4d54-9879-b8591ff6194c-grpc-us-west2-cf.aws-staging.edgedelta.com:4317 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Using HTTPS
./telemetrygen metrics \
  --otlp-endpoint https://localhost:4318 \
  --service my-service \
  --rate 5 \
  --duration 30s
```

### Logs

Generate log data:

```bash
# Using gRPCs
./telemetrygen logs \
  --otlp-endpoint grpcs://852e95f4-5bbc-4d54-9879-b8591ff6194c-grpc-us-west2-cf.aws-staging.edgedelta.com:4317 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Using HTTP with custom port
./telemetrygen logs \
  --otlp-endpoint http://localhost:8080 \
  --service my-service \
  --rate 20 \
  --duration 5s
```

## Docker Usage

```bash
# Traces
docker run --rm telemetrygen traces \
  --otlp-endpoint grpcs://your-endpoint:4317 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Metrics
docker run --rm telemetrygen metrics \
  --otlp-endpoint grpcs://your-endpoint:4317 \
  --service my-service \
  --rate 1 \
  --duration 10s

# Logs
docker run --rm telemetrygen logs \
  --otlp-endpoint grpcs://your-endpoint:4317 \
  --service my-service \
  --rate 1 \
  --duration 10s
```

## Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `--otlp-endpoint` | OTLP endpoint URL (grpc://, grpcs://, http://, https://) | - | Yes |
| `--service` | Service name for telemetry | telemetrygen | No |
| `--rate` | Number of telemetry items per second | 1 | No |
| `--duration` | How long to generate telemetry (e.g., 10s, 1m, 1h) | 10s | No |
| `--insecure` | Use insecure connection (not implemented yet) | false | No |

## Protocol Support

- `grpc://` - Insecure gRPC (default port: 4317)
- `grpcs://` - Secure gRPC with TLS (default port: 4317)
- `http://` - Insecure HTTP (default port: 4318)
- `https://` - Secure HTTP with TLS (default port: 4318)

## Default Ports

If you don't specify a port in the endpoint URL, the following defaults are used:

- gRPC/gRPCs: 4317
- HTTP/HTTPS: 4318

## Examples

```bash
# Quick test with local collector
./telemetrygen traces --otlp-endpoint grpc://localhost --service test-app --duration 5s

# High-volume test
./telemetrygen metrics --otlp-endpoint http://localhost:4318 --service load-test --rate 100 --duration 1m

# Production endpoint test
./telemetrygen logs --otlp-endpoint grpcs://prod.example.com:4317 --service prod-app --rate 10 --duration 30s
```

## What Gets Generated

### Traces
- Parent spans with child spans
- Random operation types and IDs
- Realistic timing and nesting

### Metrics
- Counter: `telemetrygen.requests`
- Histogram: `telemetrygen.duration`
- Gauge: `telemetrygen.cpu_usage`

### Logs
- Various log levels (INFO, WARN, ERROR, DEBUG)
- Realistic log messages
- Additional attributes (component, request_id, user_id)
