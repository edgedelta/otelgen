package main

import (
	"fmt"
	"os"

	"github.com/edgedelta/otelgen/pkg/otelgen"
	"github.com/spf13/cobra"
)

func init() {
	// Enable OTEL SDK debug logging if verbose
	// Users can also set OTEL_LOG_LEVEL=debug
	if os.Getenv("OTEL_LOG_LEVEL") == "" && verbose {
		os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "99")
		os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
	}
}

var (
	otlpEndpoint  string
	serviceName   string
	rate          int
	duration      string
	headers       map[string]string
	verbose       bool
	insecureSkip  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "otelgen",
		Short: "Generate telemetry data (traces, metrics, logs) for testing OTEL endpoints",
	}

	// Common flags for all commands
	addCommonFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&otlpEndpoint, "otlp-endpoint", "", "OTLP endpoint (e.g., grpcs://host:4317, http://host:4318)")
		cmd.Flags().StringVar(&serviceName, "service", "otelgen", "Service name")
		cmd.Flags().IntVar(&rate, "rate", 1, "Rate of telemetry generation per second")
		cmd.Flags().StringVar(&duration, "duration", "10s", "Duration to generate telemetry (e.g., 10s, 1m)")
		cmd.Flags().StringToStringVar(&headers, "headers", nil, "Additional headers (e.g., key1=value1,key2=value2)")
		cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
		cmd.Flags().BoolVar(&insecureSkip, "insecure-skip-verify", false, "Skip TLS certificate verification (insecure)")
		cmd.MarkFlagRequired("otlp-endpoint")
	}

	// Traces command
	tracesCmd := &cobra.Command{
		Use:   "traces",
		Short: "Generate trace data",
		RunE:  runTraces,
	}
	addCommonFlags(tracesCmd)

	// Metrics command
	metricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Generate metrics data",
		RunE:  runMetrics,
	}
	addCommonFlags(metricsCmd)

	// Logs command
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Generate log data",
		RunE:  runLogs,
	}
	addCommonFlags(logsCmd)

	rootCmd.AddCommand(tracesCmd, metricsCmd, logsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTraces(cmd *cobra.Command, args []string) error {
	endpoint, err := otelgen.ParseEndpoint(otlpEndpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	if verbose {
		fmt.Printf("Endpoint: %s\n", endpoint.String())
		fmt.Printf("Service: %s\n", serviceName)
		fmt.Printf("Rate: %d/s\n", rate)
		fmt.Printf("Duration: %s\n", duration)
		fmt.Printf("Secure: %v\n", endpoint.Secure)
		fmt.Printf("Protocol: %s\n", endpoint.Protocol)
		fmt.Printf("Insecure Skip Verify: %v\n", insecureSkip)
		if len(headers) > 0 {
			fmt.Printf("Headers: %v\n", headers)
		}
		fmt.Println()
	}

	fmt.Printf("Generating traces to %s for service %s at %d/s for %s\n",
		endpoint.String(), serviceName, rate, duration)

	return otelgen.GenerateTraces(endpoint, serviceName, rate, duration, headers, verbose, insecureSkip)
}

func runMetrics(cmd *cobra.Command, args []string) error {
	endpoint, err := otelgen.ParseEndpoint(otlpEndpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	if verbose {
		fmt.Printf("Endpoint: %s\n", endpoint.String())
		fmt.Printf("Service: %s\n", serviceName)
		fmt.Printf("Rate: %d/s\n", rate)
		fmt.Printf("Duration: %s\n", duration)
		fmt.Printf("Secure: %v\n", endpoint.Secure)
		fmt.Printf("Protocol: %s\n", endpoint.Protocol)
		fmt.Printf("Insecure Skip Verify: %v\n", insecureSkip)
		if len(headers) > 0 {
			fmt.Printf("Headers: %v\n", headers)
		}
		fmt.Println()
	}

	fmt.Printf("Generating metrics to %s for service %s at %d/s for %s\n",
		endpoint.String(), serviceName, rate, duration)

	return otelgen.GenerateMetrics(endpoint, serviceName, rate, duration, headers, verbose, insecureSkip)
}

func runLogs(cmd *cobra.Command, args []string) error {
	endpoint, err := otelgen.ParseEndpoint(otlpEndpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	if verbose {
		fmt.Printf("Endpoint: %s\n", endpoint.String())
		fmt.Printf("Service: %s\n", serviceName)
		fmt.Printf("Rate: %d/s\n", rate)
		fmt.Printf("Duration: %s\n", duration)
		fmt.Printf("Secure: %v\n", endpoint.Secure)
		fmt.Printf("Protocol: %s\n", endpoint.Protocol)
		fmt.Printf("Insecure Skip Verify: %v\n", insecureSkip)
		if len(headers) > 0 {
			fmt.Printf("Headers: %v\n", headers)
		}
		fmt.Println()
	}

	fmt.Printf("Generating logs to %s for service %s at %d/s for %s\n",
		endpoint.String(), serviceName, rate, duration)

	return otelgen.GenerateLogs(endpoint, serviceName, rate, duration, headers, verbose, insecureSkip)
}
