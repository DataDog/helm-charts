package datadog

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// systemProbeSeccompSyscalls renders the system-probe seccomp profile with
// values and returns the syscalls it allows, whatever their arguments.
func systemProbeSeccompSyscalls(t *testing.T, values map[string]string) map[string]bool {
	t.Helper()
	overrides := map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}
	for k, v := range values {
		overrides[k] = v
	}
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog", ChartPath: "../../charts/datadog",
		ShowOnly: []string{"templates/system-probe-configmap.yaml"}, Overrides: overrides,
	})
	require.NoError(t, err)
	var cm corev1.ConfigMap
	for _, doc := range strings.Split(manifest, "---") {
		if strings.Contains(doc, "name: datadog-security\n") {
			common.Unmarshal(t, doc, &cm)
			break
		}
	}
	require.Contains(t, cm.Data, "system-probe-seccomp.json")
	var profile struct {
		Syscalls []struct{ Names []string }
	}
	require.NoError(t, json.Unmarshal([]byte(cm.Data["system-probe-seccomp.json"]), &profile))
	allowed := map[string]bool{}
	for _, s := range profile.Syscalls {
		for _, n := range s.Names {
			allowed[n] = true
		}
	}
	return allowed
}

// TestSystemProbeSeccompHostBenchmarks checks that the system-probe seccomp
// profile allows the syscalls of host benchmarks when they run in
// system-probe: OpenSCAP enters the host root with chroot and fchdir, and
// libgcrypt locks its secure memory with mlock in FIPS mode.
func TestSystemProbeSeccompHostBenchmarks(t *testing.T) {
	syscalls := []string{"chroot", "fchdir", "mlock", "munlock"}
	for _, tc := range []struct {
		name   string
		values map[string]string
		want   bool
	}{
		{
			name: "host benchmarks in system-probe",
			values: map[string]string{
				"datadog.securityAgent.compliance.enabled":          "true",
				"datadog.securityAgent.compliance.runInSystemProbe": "true",
			},
			want: true,
		},
		{
			name: "xccdf in system-probe",
			values: map[string]string{
				"datadog.securityAgent.compliance.enabled":                 "true",
				"datadog.securityAgent.compliance.runInSystemProbe":        "true",
				"datadog.securityAgent.compliance.host_benchmarks.enabled": "false",
				"datadog.securityAgent.compliance.xccdf.enabled":           "true",
			},
			want: true,
		},
		{
			name: "host benchmarks disabled",
			values: map[string]string{
				"datadog.securityAgent.compliance.enabled":                 "true",
				"datadog.securityAgent.compliance.runInSystemProbe":        "true",
				"datadog.securityAgent.compliance.host_benchmarks.enabled": "false",
			},
			want: false,
		},
		{
			name: "compliance in security-agent",
			values: map[string]string{
				"datadog.securityAgent.compliance.enabled": "true",
				"datadog.networkMonitoring.enabled":        "true",
			},
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			allowed := systemProbeSeccompSyscalls(t, tc.values)
			for _, s := range syscalls {
				assert.Equal(t, tc.want, allowed[s], s)
			}
		})
	}
}
