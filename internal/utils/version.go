package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// VersionComparisonMode defines how to handle version comparisons
type VersionComparisonMode string

const (
	VersionModeNumeric     VersionComparisonMode = "numeric"     // Convert to numeric (e.g., 5.0.0 -> 5000000)
	VersionModeLexographic VersionComparisonMode = "lexographic" // Use string comparison (default DQL)
	VersionModeSemantic    VersionComparisonMode = "semantic"    // Parse semantic versions
)

// VersionField represents a field that contains version information
type VersionField struct {
	Field string                `json:"field"`
	Mode  VersionComparisonMode `json:"mode"`
}

// ConvertVersionToNumeric converts a semantic version string to a comparable number
// Example: "5.0.0" -> 5000000, "10.2.1" -> 10002001
func ConvertVersionToNumeric(version string) (int64, error) {
	// Remove any non-numeric prefixes (like 'v')
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) < 1 || len(parts) > 4 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	// Pad with zeros if needed
	for len(parts) < 3 {
		parts = append(parts, "0")
	}

	var result int64
	multipliers := []int64{1000000, 1000, 1}

	for i, part := range parts[:3] {
		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid version part '%s' in %s", part, version)
		}

		if num > 999 {
			return 0, fmt.Errorf("version part %d too large in %s", num, version)
		}

		result += num * multipliers[i]
	}

	return result, nil
}

// IsVersionField checks if a field should be treated as a version field
func IsVersionField(fieldName string) bool {
	versionFields := []string{
		"app_version",
		"os_version",
		"version",
		"api_version",
		"client_version",
	}

	for _, vf := range versionFields {
		if strings.Contains(strings.ToLower(fieldName), vf) {
			return true
		}
	}

	return false
}

// GetVersionComparisonMode returns the appropriate comparison mode for a version field
func GetVersionComparisonMode(fieldName string) VersionComparisonMode {
	// For now, default to numeric conversion for known version fields
	if IsVersionField(fieldName) {
		return VersionModeNumeric
	}
	return VersionModeLexographic
}
