package yamlmapper

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	agentConfStrictEnv      = "YAMLMAPPER_AGENT_CONF_STRICT" // enable strict agent config comparison
	staleCleanupEnabledEnv  = "YAMLMAPPER_CLEANUP_STALE"     // enable stale resource cleanup
	mapperWarningsStrictEnv = "YAMLMAPPER_WARNINGS_STRICT"   // enable strict mapper warnings

	ddaCRDName              = "datadogagents.datadoghq.com"
)

// requiredCRDs lists the CRDs that must be present for integration tests to run
var requiredCRDs = []string{
	"datadogagents.datadoghq.com",
	"datadogagentinternals.datadoghq.com",
}

// =============================================================================
// Environment and configuration
// =============================================================================

func validateEnv(t *testing.T) {
	// Check cluster context is not production
	context := common.CurrentContext(t)
	t.Log("Checking current context:", context)
	if strings.Contains(strings.ToLower(context), "staging") ||
		strings.Contains(strings.ToLower(context), "prod") {
		t.Fatal("Make sure context is pointing to local cluster")
	}

	// Check required CRDs are installed
	t.Log("Checking required CRDs are installed...")
	kubectlOptions := k8s.NewKubectlOptions("", "", "")

	var missingCRDs []string
	for _, crd := range requiredCRDs {
		_, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, "get", "crd", crd)
		if err != nil {
			missingCRDs = append(missingCRDs, crd)
		}
	}

	if len(missingCRDs) > 0 {
		t.Fatalf(`Required CRDs not found: %v

To install the required CRDs, run:
  helm install datadog-crds ./charts/datadog-crds \
    --create-namespace --namespace datadog-crds \
    --set crds.datadogAgents=true \
    --set crds.datadogAgentInternals=true

Or use the Makefile target:
  make setup-mapper-crds
`, missingCRDs)
	}
	t.Log("All required CRDs are present")
}

func isAgentConfStrict() bool {
	return strings.EqualFold(os.Getenv(agentConfStrictEnv), "1") ||
		strings.EqualFold(os.Getenv(agentConfStrictEnv), "true") ||
		strings.EqualFold(os.Getenv(agentConfStrictEnv), "yes")
}

// isMapperWarningsStrict returns true if mapper warnings should cause test failures
func isMapperWarningsStrict() bool {
	return strings.EqualFold(os.Getenv(mapperWarningsStrictEnv), "1") ||
		strings.EqualFold(os.Getenv(mapperWarningsStrictEnv), "true") ||
		strings.EqualFold(os.Getenv(mapperWarningsStrictEnv), "yes")
}

// =============================================================================
// Agent config normalization
// =============================================================================

func expectedDsCount(t *testing.T, kubectlOptions *k8s.KubectlOptions) int {
	nodes := k8s.GetNodes(t, kubectlOptions)
	cpNodes, _ := k8s.GetNodesByFilterE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/control-plane"})

	dsCount := len(nodes) - len(cpNodes)
	if dsCount == 0 {
		// Some local clusters schedule on control-plane nodes; avoid undercounting.
		return len(nodes)
	}
	return dsCount
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
	normalizeDataWithPath(m, "")
}

func normalizeDataWithPath(m map[string]interface{}, parentPath string) {
	for k, v := range m {
		currentPath := k
		if parentPath != "" {
			currentPath = parentPath + "." + k
		}

		// Check if this key or full path should be skipped
		if _, ok := skipFields[k]; ok {
			delete(m, k)
			continue
		}
		if _, ok := skipFields[currentPath]; ok {
			delete(m, k)
			continue
		}

		switch val := v.(type) {
		case bool:
			m[k] = fmt.Sprintf("%v", val)
		case map[string]interface{}:
			normalizeDataWithPath(val, currentPath)
		}
	}
}

// isTimestampLine checks if a line starts with a timestamp in the format "2006-01-02 15:04:05 UTC"
func isTimestampLine(line string) bool {
	// 23 = length of "2006-01-02 15:04:05 UTC" (date + space + time + space + timezone)
	const timestampLength = 23
	if len(line) < timestampLength {
		return false
	}
	_, err := time.Parse("2006-01-02 15:04:05 MST", line[:timestampLength])
	return err == nil
}

// skipFields is a set of field names or dot-separated paths that should be skipped during comparison.
// Use simple field names for top-level keys, or full paths for nested fields (e.g., "parent.child.field").
var skipFields = map[string]struct{}{
	"install_id":              {},
	"install_time":            {},
	"install_type":            {},
	"kubernetes_service_name": {},
	"kubernetes_kubelet_host": {},
	"token_name":              {},
	"site":                    {},
	"app_key":                 {},
	"expvar_port":             {},
	"log_level":               {},
	// Nested paths
	"orchestrator_explorer.kubelet_config_check.enabled": {}, // TODO: remove this when available in operator

	// kubelet_client_ca: Behavioral disparity between Helm and Operator
	// - Helm: Sets DD_KUBELET_CLIENT_CA when agentCAPath is provided (allows referencing existing files like k8s service account CA)
	// - Operator: Only sets DD_KUBELET_CLIENT_CA when hostCAPath is provided (assumes all CA paths need explicit host mounts)
	// See: datadog-operator/internal/controller/datadogagent/global/agent.go lines 39-75
	// TODO: Consider filing operator enhancement to support agentCAPath without hostCAPath
	"kubelet_client_ca": {},
}

// Label selectors for different agent installation types
const (
	helmAgentLabelSelector        = "app.kubernetes.io/component=agent,app.kubernetes.io/managed-by=Helm"
	helmClusterAgentLabelSelector = "app.kubernetes.io/component=cluster-agent,app.kubernetes.io/managed-by=Helm"
	helmCCRLabelSelector          = "app.kubernetes.io/component=clusterchecks-agent,app.kubernetes.io/managed-by=Helm"

	operatorAgentLabelSelector        = "agent.datadoghq.com/component=agent,app.kubernetes.io/managed-by=datadog-operator"
	operatorClusterAgentLabelSelector = "agent.datadoghq.com/component=cluster-agent,app.kubernetes.io/managed-by=datadog-operator"
	operatorCCRLabelSelector          = "agent.datadoghq.com/component=cluster-checks-runner,app.kubernetes.io/managed-by=datadog-operator"
)

// getAgentConf retrieves the agent config from an agent pod matching the given label selector.
// It waits for the pod to be available, executes 'agent config --all', and normalizes the output.
func getAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, retries int) string {
	pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	})
	require.NotEmpty(t, pods, "No agent pods found with selector: %s", labelSelector)

	podName := pods[0].Name
	err := k8s.WaitUntilPodAvailableE(t, kubectlOptions, podName, retries, 15*time.Second)
	if err != nil {
		t.Logf("Failed to wait for agent pod %s: %v", podName, err)
		return ""
	}

	agentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, "exec", podName, "--", "agent", "config", "--all")
	if err != nil {
		t.Logf("Failed to get agent config from pod %s: %v", podName, err)
		return ""
	}
	return normalizeAgentConf(agentConf)
}

// getHelmAgentConf retrieves the agent config from a helm-installed agent pod
func getHelmAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions) string {
	return getAgentConf(t, kubectlOptions, helmAgentLabelSelector, 10)
}

// getOperatorAgentConf retrieves the agent config from an operator-installed agent pod
func getOperatorAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions) string {
	return getAgentConf(t, kubectlOptions, operatorAgentLabelSelector, 5)
}

// =============================================================================
// Log Capture
// =============================================================================

// MapperLogSink captures slog output during mapper runs.
// It is safe for concurrent use from multiple goroutines.
type MapperLogSink struct {
	mu       sync.RWMutex
	warnings []string
	errors   []string
	infos    []string
}

// newMapperLogSink creates a new log capture instance
func newMapperLogSink() *MapperLogSink {
	return &MapperLogSink{
		warnings: make([]string, 0, 8),
		errors:   make([]string, 0, 8),
		infos:    make([]string, 0, 16),
	}
}

// Enabled implements slog.Handler
func (c *MapperLogSink) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

// Handle implements slog.Handler - thread-safe log message capture
func (c *MapperLogSink) Handle(_ context.Context, r slog.Record) error {
	// Build the message with attributes
	var sb strings.Builder
	sb.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		return true
	})
	msg := sb.String()

	c.mu.Lock()
	defer c.mu.Unlock()

	switch r.Level {
	case slog.LevelWarn:
		c.warnings = append(c.warnings, msg)
	case slog.LevelError:
		c.errors = append(c.errors, msg)
	case slog.LevelInfo:
		c.infos = append(c.infos, msg)
	}
	return nil
}

// WithAttrs implements slog.Handler
func (c *MapperLogSink) WithAttrs(attrs []slog.Attr) slog.Handler {
	return c
}

// WithGroup implements slog.Handler
func (c *MapperLogSink) WithGroup(name string) slog.Handler {
	return c
}

// GetWarnings returns a copy of all captured warnings (thread-safe)
func (c *MapperLogSink) GetWarnings() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.warnings))
	copy(result, c.warnings)
	return result
}

// GetErrors returns a copy of all captured errors (thread-safe)
func (c *MapperLogSink) GetErrors() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.errors))
	copy(result, c.errors)
	return result
}

// GetInfos returns a copy of all captured info messages (thread-safe)
func (c *MapperLogSink) GetInfos() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.infos))
	copy(result, c.infos)
	return result
}

// HasWarnings returns true if any warnings were captured (thread-safe)
func (c *MapperLogSink) HasWarnings() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.warnings) > 0
}

// HasErrors returns true if any errors were captured (thread-safe)
func (c *MapperLogSink) HasErrors() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.errors) > 0
}

// WarningCount returns the number of captured warnings (thread-safe)
func (c *MapperLogSink) WarningCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.warnings)
}

// ErrorCount returns the number of captured errors (thread-safe)
func (c *MapperLogSink) ErrorCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.errors)
}

// Clear resets all captured logs (thread-safe)
func (c *MapperLogSink) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.warnings = c.warnings[:0]
	c.errors = c.errors[:0]
	c.infos = c.infos[:0]
}

// captureMapperLogs installs the log capture as the default slog handler and returns
// the capture instance and a cleanup function to restore the original handler.
// Tests are run sequentially, so we do not guard the global handler swap with a mutex.
func captureMapperLogs() (*MapperLogSink, func()) {
	capture := newMapperLogSink()
	originalHandler := slog.Default().Handler()
	slog.SetDefault(slog.New(capture))

	return capture, func() {
		slog.SetDefault(slog.New(originalHandler))
	}
}

// reportMapperWarnings logs warnings and optionally fails the test if strict mode is enabled
func reportMapperWarnings(t *testing.T, capture *MapperLogSink) {
	t.Helper()
	warnings := capture.GetWarnings()
	if len(warnings) > 0 {
		t.Logf("Mapper warnings (%d):", len(warnings))
		for _, w := range warnings {
			t.Logf("  WARN: %s", w)
		}
		if isMapperWarningsStrict() {
			t.Fatalf("Strict mode: mapper produced %d warning(s)", len(warnings))
		}
	}
}

// reportMapperErrors logs errors captured during mapper execution
func reportMapperErrors(t *testing.T, capture *MapperLogSink) {
	t.Helper()
	errors := capture.GetErrors()
	if len(errors) > 0 {
		t.Logf("Mapper errors (%d):", len(errors))
		for _, e := range errors {
			t.Logf("  ERROR: %s", e)
		}
	}
}

// =============================================================================
// Kubernetes wait helpers
// =============================================================================

// waitForPodsTerminated waits until all pods matching the label selector are fully terminated
// This helps prevent containerd state corruption from rapid pod creation/deletion
func waitForPodsTerminated(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, timeout time.Duration) error {
	t.Logf("Waiting for pods with selector '%s' to terminate...", labelSelector)

	// Calculate max retries: timeout / sleep interval
	sleepInterval := 2 * time.Second
	maxRetries := int(timeout / sleepInterval)

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for pods with selector '%s' to terminate", labelSelector),
		maxRetries, sleepInterval, func() (string, error) {
			pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if len(pods) == 0 {
				t.Log("All pods terminated successfully")
				return "", nil
			}
			return "", fmt.Errorf("still waiting for %d pods to terminate", len(pods))
		})
	return err
}

// interTestDelay adds a small delay between tests to allow containerd to stabilize
// This helps prevent containerd state corruption from rapid container operations
func interTestDelay(t *testing.T, duration time.Duration) {
	t.Logf("Waiting %v between tests for containerd stability...", duration)
	time.Sleep(duration)
}

// waitForNamespaceDeletion waits for a namespace to be fully deleted from the cluster.
// If the namespace is stuck terminating, it will attempt to force delete by removing finalizers.
func waitForNamespaceDeletion(t *testing.T, namespace string, timeout time.Duration) {
	t.Logf("Waiting for namespace %s to be fully deleted", namespace)
	// Use a kubectlOptions without namespace for cluster-scoped operations
	kubectlOptions := k8s.NewKubectlOptions("", "", "")

	// Calculate max retries: timeout / sleep interval
	sleepInterval := 5 * time.Second
	maxRetries := int(timeout / sleepInterval)
	forceDeleteAttempted := false
	retryCount := 0

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for namespace %s to be deleted", namespace),
		maxRetries, sleepInterval, func() (string, error) {
			retryCount++

			// Check if namespace exists
			output, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, "get", "namespace", namespace, "-o", "name")
			if err != nil || strings.TrimSpace(output) == "" {
				t.Logf("Namespace %s has been deleted", namespace)
				return "", nil
			}

			// Check if namespace is stuck terminating
			phase, _ := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, "get", "namespace", namespace, "-o", "jsonpath={.status.phase}")
			if strings.TrimSpace(phase) == "Terminating" {
				t.Logf("Namespace %s is terminating, waiting...", namespace)
				// If stuck terminating for a while (>70% of timeout), try to force delete by removing finalizers
				if !forceDeleteAttempted && retryCount > int(float64(maxRetries)*0.7) {
					t.Logf("Attempting to force delete stuck namespace %s by removing finalizers", namespace)
					forceDeleteAttempted = true

					// Remove finalizers from namespace
					_ = k8s.RunKubectlE(t, kubectlOptions, "patch", "namespace", namespace, "--type=merge", "-p", `{"spec":{"finalizers":null}}`)
					_ = k8s.RunKubectlE(t, kubectlOptions, "patch", "namespace", namespace, "--type=merge", "-p", `{"metadata":{"finalizers":null}}`)
				}
			}

			return "", fmt.Errorf("namespace %s still exists", namespace)
		})

	if err != nil {
		// Final force delete attempt
		t.Logf("Warning: Timeout waiting for namespace %s deletion, attempting final force delete", namespace)
		_ = k8s.RunKubectlE(t, kubectlOptions, "delete", "namespace", namespace, "--force", "--grace-period=0", "--ignore-not-found")
	}
}

// waitForDDADeletion waits for all DatadogAgent resources in a namespace to be fully deleted.
// It handles stuck resources by removing finalizers if necessary.
func waitForDDADeletion(t *testing.T, kubectlOptions *k8s.KubectlOptions, namespace string, timeout time.Duration) {
	t.Logf("Waiting for DDA resources to be deleted in namespace %s", namespace)

	// Calculate max retries: timeout / sleep interval
	sleepInterval := 5 * time.Second
	maxRetries := int(timeout / sleepInterval)

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for DDA resources to be deleted in namespace %s", namespace),
		maxRetries, sleepInterval, func() (string, error) {
			// Check if any DatadogAgent resources exist in the namespace
			output, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, "get", ddaCRDName, "-o", "name")
			if err != nil || strings.TrimSpace(output) == "" {
				// No resources found or error (likely means no resources)
				t.Logf("No DDA resources found in namespace %s", namespace)
				return "", nil
			}

			// Resources still exist - check if they're stuck with finalizers
			resources := strings.Split(strings.TrimSpace(output), "\n")
			for _, resource := range resources {
				if resource == "" {
					continue
				}
				// Extract just the name from "datadogagent.datadoghq.com/name"
				parts := strings.Split(resource, "/")
				name := parts[len(parts)-1]

				// Check if resource is stuck in terminating state
				status, _ := k8s.RunKubectlAndGetOutputE(t, kubectlOptions,
					"get", ddaCRDName, name, "-o", "jsonpath={.metadata.deletionTimestamp}")
				if status != "" {
					// Resource is terminating but stuck - remove finalizers
					t.Logf("DDA %s is stuck terminating, removing finalizers", name)
					_ = k8s.RunKubectlE(t, kubectlOptions,
						"patch", ddaCRDName, name, "--type=merge", "-p", `{"metadata":{"finalizers":null}}`)
				}
			}

			return "", fmt.Errorf("%d DDA resources still exist in namespace %s", len(resources), namespace)
		})

	if err != nil {
		t.Logf("Warning: Timeout waiting for DDA deletion in namespace %s", namespace)
	}
}

// =============================================================================
// Stale resource cleanup
// =============================================================================

// cleanupStaleResources removes any leftover test namespaces from previous runs.
// This handles the case where tests were interrupted with Ctrl+C and t.Cleanup() didn't run.
// It uses force-deletion since these are orphaned resources that need quick cleanup.
func cleanupStaleResources() {
	log.Printf("Checking for stale test resources from previous runs...")

	if !isStaleCleanupEnabled() {
		log.Printf("Stale cleanup disabled (set %s=true to enable)", staleCleanupEnabledEnv)
		return
	}

	if !isSafeContext() {
		log.Printf("Stale cleanup skipped: unsafe kubectl context")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	staleNamespaces := findStaleNamespaces(ctx)
	if len(staleNamespaces) == 0 {
		log.Printf("No stale test namespaces found")
		return
	}

	log.Printf("Found %d stale test namespace(s): %v", len(staleNamespaces), staleNamespaces)

	for _, ns := range staleNamespaces {
		forceDeleteNamespace(ctx, ns)
	}

	log.Printf("Stale resource cleanup complete")
}

// findStaleNamespaces returns a list of test namespaces that exist in the cluster
func findStaleNamespaces(ctx context.Context) []string {
	cmd := exec.CommandContext(ctx, "kubectl", "get", "namespaces", "-o", "name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: Could not list namespaces: %v", err)
		return nil
	}

	var staleNamespaces []string
	for _, ns := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		name := strings.TrimPrefix(ns, "namespace/")
		if strings.HasPrefix(name, testNamespacePrefix) {
			staleNamespaces = append(staleNamespaces, name)
		}
	}
	return staleNamespaces
}

func isStaleCleanupEnabled() bool {
	return strings.EqualFold(os.Getenv(staleCleanupEnabledEnv), "1") ||
		strings.EqualFold(os.Getenv(staleCleanupEnabledEnv), "true") ||
		strings.EqualFold(os.Getenv(staleCleanupEnabledEnv), "yes")
}

func isSafeContext() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "config", "current-context")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: Could not determine current context: %v", err)
		return false
	}

	contextName := strings.ToLower(strings.TrimSpace(string(output)))
	if contextName == "" {
		return false
	}

	if strings.Contains(contextName, "staging") || strings.Contains(contextName, "prod") {
		return false
	}

	return true
}

// forceDeleteNamespace aggressively removes a namespace and all its resources.
// This is used for cleanup of orphaned test namespaces where we don't care about graceful shutdown.
func forceDeleteNamespace(ctx context.Context, namespace string) {
	log.Printf("Force-deleting stale namespace: %s", namespace)

	// Remove finalizers from all DDAs (allows immediate deletion)
	removeDDAFinalizersInNamespace(ctx, namespace)

	// Delete DDAs
	runKubectl(ctx, "delete", ddaCRDName, "--all", "-n", namespace, "--ignore-not-found", "--timeout=10s")

	// Uninstall helm releases
	for _, release := range []string{releaseDatadogOperator, releaseDatadog} {
		runHelm(ctx, "uninstall", release, "-n", namespace, "--ignore-not-found", "--timeout=30s")
	}

	// Delete namespace (with force fallback)
	if !runKubectl(ctx, "delete", "namespace", namespace, "--ignore-not-found", "--timeout=30s") {
		log.Printf("  Forcing namespace deletion for %s...", namespace)
		runKubectl(ctx, "patch", "namespace", namespace, "--type=merge", "-p", `{"spec":{"finalizers":null},"metadata":{"finalizers":null}}`)
		runKubectl(ctx, "delete", "namespace", namespace, "--force", "--grace-period=0", "--ignore-not-found")
	}

	// Brief wait for deletion
	waitForNamespaceGone(ctx, namespace, 30*time.Second)
}

// removeDDAFinalizersInNamespace removes finalizers from all DDAs in a namespace
func removeDDAFinalizersInNamespace(ctx context.Context, namespace string) {
	cmd := exec.CommandContext(ctx, "kubectl", "get", ddaCRDName,
		"-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	output, _ := cmd.CombinedOutput()

	for _, name := range strings.Fields(string(output)) {
		runKubectl(ctx, "patch", ddaCRDName, name,
			"-n", namespace, "--type=merge", "-p", `{"metadata":{"finalizers":null}}`)
	}
}

// waitForNamespaceGone waits for a namespace to be fully deleted
func waitForNamespaceGone(ctx context.Context, namespace string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(ctx, "kubectl", "get", "namespace", namespace, "-o", "name")
		output, err := cmd.CombinedOutput()
		if err != nil || strings.TrimSpace(string(output)) == "" {
			log.Printf("  Namespace %s deleted successfully", namespace)
			return
		}
		time.Sleep(5 * time.Second)
	}
	log.Printf("  Warning: Namespace %s may still be terminating", namespace)
}

// runKubectl executes a kubectl command and returns true if successful
func runKubectl(ctx context.Context, args ...string) bool {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Don't log "not found" errors
		if !strings.Contains(string(output), "not found") && !strings.Contains(string(output), "NotFound") {
			log.Printf("  kubectl %v: %v", args[0], err)
		}
		return false
	}
	return true
}

// runHelm executes a helm command
func runHelm(ctx context.Context, args ...string) bool {
	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "not found") {
			log.Printf("  helm %v: %v", args[0], err)
		}
		return false
	}
	return true
}
