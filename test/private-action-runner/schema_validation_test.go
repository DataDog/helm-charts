package private_action_runner

import (
	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/assert"
	"sort"
	"strings"
	"testing"
)

func TestPreviousValuesSchema(t *testing.T) {

	command := common.HelmCommand{
		ReleaseName: "0-dot-x-values-file",
		ChartPath:   "../../charts/private-action-runner",
		Values:      []string{"./data/old-values-file.yaml"},
	}
	expectedErrorLines := []string{
		"- (root): Additional property credentialFiles is not allowed",
		"- (root): Additional property credentialSecrets is not allowed",
		"- (root): Additional property runners is not allowed",
		"Error: values don't meet the specifications of the schema(s) in the following chart(s):",
		"private-action-runner:",
	}

	_, err := common.RenderChart(t, command)
	assert.NotNil(t, err, "expect schema validation error")
	stderr := err.(*shell.ErrWithCmdOutput).Output.Stderr()
	cleanedStderr := uniqueSortedLines(strings.Split(dropLinesStartingWith(stderr, "install.go:", "helm.go:"), "\n"))
	assert.Equal(t, expectedErrorLines, cleanedStderr)
}

func dropLinesStartingWith(input string, prefixes ...string) string {
	var result strings.Builder
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		shouldDrop := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				shouldDrop = true
				break
			}
		}
		if !shouldDrop {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}
func uniqueSortedLines(lines []string) []string {
	lineMap := make(map[string]struct{})
	for _, line := range lines {
		if line != "" {
			lineMap[line] = struct{}{}
		}
	}
	uniqueLines := make([]string, 0, len(lineMap))
	for line := range lineMap {
		uniqueLines = append(uniqueLines, line)
	}
	sort.Strings(uniqueLines)
	return uniqueLines
}
