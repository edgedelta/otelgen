package otelgen

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSize parses a size string like "1kb", "1mb", "500b" into bytes
func ParseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))

	// Extract number and unit
	var numStr string
	var unit string

	for i, c := range sizeStr {
		if c >= '0' && c <= '9' || c == '.' {
			numStr += string(c)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %w", err)
	}

	var multiplier int64
	switch unit {
	case "b", "":
		multiplier = 1
	case "kb", "k":
		multiplier = 1024
	case "mb", "m":
		multiplier = 1024 * 1024
	case "gb", "g":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit: %s (supported: b, kb, mb, gb)", unit)
	}

	return int64(num * float64(multiplier)), nil
}

// GeneratePadding creates a padding string of the specified size
func GeneratePadding(size int64) string {
	if size <= 0 {
		return ""
	}
	// Create a buffer of the specified size filled with 'x' characters
	padding := make([]byte, size)
	for i := range padding {
		padding[i] = 'x'
	}
	return string(padding)
}
