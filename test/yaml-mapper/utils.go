//go:build integration

package yaml_mapper

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CleanupRegistry stores test artifacts that require cleanup after each test run. `files` stores mapped DDA manifest files. `datadog` stores the datadog helm chart uninstall function. `operator` stores the operator chart uninstall function.
type CleanupRegistry struct {
	mu       sync.Mutex
	files    []string
	datadog  func()
	operator func()
}

func (d *CleanupRegistry) AddDDA(files ...string) {
	d.mu.Lock()
	d.files = append(d.files, files...)
	d.mu.Unlock()
}

func (d *CleanupRegistry) AddDatadog(cleanup func()) {
	d.mu.Lock()
	d.datadog = cleanup
	d.mu.Unlock()
}

func (d *CleanupRegistry) UnsetDatadog() {
	d.mu.Lock()
	d.datadog = nil
	d.mu.Unlock()
}

func (d *CleanupRegistry) AddOperator(cleanup func()) {
	d.mu.Lock()
	d.operator = cleanup
	d.mu.Unlock()
}

func (d *CleanupRegistry) UnsetOperator() {
	d.mu.Lock()
	d.operator = nil
	d.mu.Unlock()
}

func (d *CleanupRegistry) GetFiles() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	cp := make([]string, len(d.files))
	copy(cp, d.files)
	return cp
}

// getHelmReleaseName for a given short release name
func getHelmReleaseName(t *testing.T, kubectlOptions *k8s.KubectlOptions, namespace string, shortReleaseName string) string {
	t.Log("Finding Helm release name...")
	helmListOutput, err := helm.RunHelmCommandAndGetOutputE(t, &helm.Options{KubectlOptions: kubectlOptions}, "list", "-n", namespace, "--short")
	require.NoError(t, err, "failed to list helm releases")

	var releaseName string
	releaseNames := strings.Split(strings.TrimSpace(helmListOutput), "\n")
	for _, release := range releaseNames {
		release = strings.TrimSpace(release)
		if strings.HasPrefix(release, shortReleaseName+"-") {
			releaseName = release
			break
		}
	}
	require.NotEmpty(t, releaseName, fmt.Sprintf("could not find release %v", releaseName))
	t.Logf("Found %s release name: %s", shortReleaseName, releaseName)
	return releaseName
}

func validateEnv(t *testing.T) {
	context := common.CurrentContext(t)
	t.Log("Checking current context:", context)
	if strings.Contains(strings.ToLower(context), "staging") ||
		strings.Contains(strings.ToLower(context), "prod") {
		t.Fatal("Make sure context is pointing to local cluster")
	}
}

func expectedDsCount(t *testing.T, kubectlOptions *k8s.KubectlOptions) int {
	nodes := k8s.GetNodes(t, kubectlOptions)
	cpNodes, _ := k8s.GetNodesByFilterE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/control-plane"})

	return len(nodes) - len(cpNodes)
}

// normalizeAgentConf removes log lines that start with timestamps in the format "2006-01-02 15:04:05 UTC"
func normalizeAgentConf(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Defaulted container") {
			continue
		}

		// Skip lines that start with a timestamp
		if isTimestampLine(line) {
			continue
		}

		result.WriteString(line)
		result.WriteByte('\n')
	}

	// Normalize data by converting bools to strings and skipping unnecessary fields
	// Unmarshal the string to map[string]interface{} first
	confData := []byte(result.String())
	var confOut map[string]interface{}
	err := yaml.Unmarshal(confData, &confOut)
	if err != nil {
		log.Printf("could not unmarshal agent config: %v", err)
		return result.String()
	}
	normalizeData(confOut)

	resultData, err := yaml.Marshal(confOut)
	if err != nil {
		log.Printf("could not marshal agent config: %v", err)
		return result.String()
	}

	return string(resultData)
}

// normalizeData walks through a map[string]interface{} recursively
// and replaces any bool value with its string equivalent ("true"/"false").
// It also filters out fields that should be skipped
func normalizeData(m map[string]interface{}) {
	for k, v := range m {
		if _, ok := skipFields[k]; ok {
			delete(m, k)
		}
		switch val := v.(type) {
		case bool:
			m[k] = fmt.Sprintf("%v", val)
		case map[string]interface{}:
			normalizeData(val)
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

// skipFields fields in the agent config output that should be skipped
var skipFields = map[string]interface{}{
	"install_id":              nil,
	"install_time":            nil,
	"install_type":            nil,
	"kubernetes_service_name": nil, // service name differs according to installation
	"kubernetes_kubelet_host": nil, // may also differ
	"token_name":              nil,
	"site":                    nil,
}
