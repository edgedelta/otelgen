package otelgen

import (
	"fmt"
	"net/url"
	"strings"
)

// Protocol represents the transport protocol
type Protocol int

const (
	ProtocolGRPC Protocol = iota
	ProtocolGRPCS
	ProtocolHTTP
	ProtocolHTTPS
)

func (p Protocol) String() string {
	switch p {
	case ProtocolGRPC:
		return "grpc"
	case ProtocolGRPCS:
		return "grpcs"
	case ProtocolHTTP:
		return "http"
	case ProtocolHTTPS:
		return "https"
	default:
		return "unknown"
	}
}

// Endpoint represents a parsed OTLP endpoint
type Endpoint struct {
	Protocol Protocol
	Host     string
	Port     string
	Secure   bool
}

// String returns the full endpoint URL
func (e *Endpoint) String() string {
	return fmt.Sprintf("%s://%s:%s", e.Protocol, e.Host, e.Port)
}

// Address returns host:port
func (e *Endpoint) Address() string {
	return fmt.Sprintf("%s:%s", e.Host, e.Port)
}

// IsGRPC returns true if the protocol is gRPC-based
func (e *Endpoint) IsGRPC() bool {
	return e.Protocol == ProtocolGRPC || e.Protocol == ProtocolGRPCS
}

// IsHTTP returns true if the protocol is HTTP-based
func (e *Endpoint) IsHTTP() bool {
	return e.Protocol == ProtocolHTTP || e.Protocol == ProtocolHTTPS
}

// ParseEndpoint parses the endpoint string and returns an Endpoint
// Supports: grpc://host:port, grpcs://host:port, http://host:port, https://host:port
// If port is not specified, uses default ports: 4317 for gRPC, 4318 for HTTP
func ParseEndpoint(endpoint string) (*Endpoint, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	// Parse the URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	ep := &Endpoint{
		Host: u.Hostname(),
	}

	// Determine protocol
	switch strings.ToLower(u.Scheme) {
	case "grpc":
		ep.Protocol = ProtocolGRPC
		ep.Secure = false
	case "grpcs":
		ep.Protocol = ProtocolGRPCS
		ep.Secure = true
	case "http":
		ep.Protocol = ProtocolHTTP
		ep.Secure = false
	case "https":
		ep.Protocol = ProtocolHTTPS
		ep.Secure = true
	default:
		return nil, fmt.Errorf("unsupported protocol: %s (supported: grpc, grpcs, http, https)", u.Scheme)
	}

	// Determine port
	port := u.Port()
	if port == "" {
		// Use default ports
		if ep.IsGRPC() {
			port = "4317"
		} else {
			port = "4318"
		}
	}
	ep.Port = port

	if ep.Host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}

	return ep, nil
}
