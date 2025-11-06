//go:build integration

package yaml_mapper

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// normalizeAgentConf removes log lines that start with timestamps in the format "2006-01-02 15:04:05 UTC"
func normalizeAgentConf(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		// Skip lines that start with a timestamp
		if isTimestampLine(line) {
			continue
		}
		// Skip lines that contain fields that should be skipped
		if shouldSkipField(line) {
			continue
		}
		result.WriteString(line)
		result.WriteByte('\n')
	}

	// Normalize bool values to string
	// Unmarshal the string to map[string]interface{} first
	confData := []byte(result.String())
	var confOut map[string]interface{}
	err := yaml.Unmarshal(confData, &confOut)
	if err != nil {
		return result.String()
	}
	convertBoolsToStrings(confOut)

	resultData, err := yaml.Marshal(confOut)
	if err != nil {
		return result.String()
	}

	return string(resultData)
}

// convertBoolsToStrings walks through a map[string]interface{} recursively
// and replaces any bool value with its string equivalent ("true"/"false").
func convertBoolsToStrings(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case bool:
			m[k] = fmt.Sprintf("%v", val)
		case map[string]interface{}:
			convertBoolsToStrings(val)
		}
	}
}

// isTimestampLine checks if a line starts with a timestamp in the format "2006-01-02 15:04:05 UTC"
func isTimestampLine(line string) bool {
	if len(line) < 23 { // Minimum length for "2006-01-02 15:04:05"
		return false
	}
	_, err := time.Parse("2006-01-02 15:04:05 MST", line[:23])
	if err == nil {
		return true
	}
	return false
}

var skipFields = map[string]interface{}{
	"install_id":              nil,
	"install_time":            nil,
	"install_type":            nil,
	"kubernetes_service_name": nil, // service name differs according to installation
	"kubernetes_kubelet_host": nil, // may also differ
	"token_name":              nil,
	"site":                    nil,
}

func shouldSkipField(line string) bool {
	for field := range skipFields {
		if strings.Contains(line, field) {
			return true
		}
	}
	return false
}
