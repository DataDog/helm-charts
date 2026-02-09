package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestAPMConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		values  string
		isValid bool
	}{
		{
			name:    "valid enabled configuration",
			values:  "valid_enabled.yaml",
			isValid: true,
		},
		{
			name:    "valid target configuration",
			values:  "valid_targets.yaml",
			isValid: true,
		},
		{
			name:    "valid namespace configuration",
			values:  "valid_namespace.yaml",
			isValid: true,
		},
		{
			name:    "both namespaces and targets",
			values:  "namespaces_and_targets.yaml",
			isValid: false,
		},
		{
			name:    "both libversions and targets",
			values:  "libversions_and_targets.yaml",
			isValid: false,
		},
		{
			name:    "both enabled and disabled namespaces",
			values:  "enabled_and_disabled_namespaces.yaml",
			isValid: false,
		},
		{
			name:    "both matchLabels and matchNames for namespace selector",
			values:  "namespace_labels_and_names.yaml",
			isValid: false,
		},
		{
			name:    "both matchExpressions and matchNames for namespace selector",
			values:  "namespace_exprs_and_names.yaml",
			isValid: false,
		},
		{
			name:    "extraneous instrumentation key",
			values:  "extra_instrumentation_key.yaml",
			isValid: false,
		},
		{
			name:    "extraneous target key",
			values:  "extra_target_key.yaml",
			isValid: false,
		},
		{
			name:    "extraneous pod selector key",
			values:  "extra_podselector_key.yaml",
			isValid: false,
		},
		{
			name:    "extraneous namespace selector key",
			values:  "extra_namespaceselector_key.yaml",
			isValid: false,
		},
		{
			name:    "ddTraceConfigs and valueFrom",
			values:  "values_from.yaml",
			isValid: true,
		},
		{
			name:    "ddTraceConfigs and valueFrom invalid",
			values:  "values_from_invalid.yaml",
			isValid: false,
		},
		{
			name:    "injectionMode csi without csi.enabled",
			values:  "injection_mode_csi_without_driver.yaml",
			isValid: false,
		},
		{
			name:    "injectionMode csi with csi.enabled",
			values:  "injection_mode_csi_with_driver.yaml",
			isValid: true,
		},
		{
			name:    "injectionMode image_volume",
			values:  "injection_mode_image_volume.yaml",
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helm := common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"../../charts/datadog/values.yaml", "values/instrumentation/" + tt.values},
			}
			_, err := common.RenderChart(t, helm)
			if tt.isValid {
				assert.Nil(t, err, "expected no error, got %v", err)
			} else {
				assert.NotNil(t, err, "expected error, got nil")
			}
		})
	}
}

const ddApmInjectionModeEnvVar = "DD_APM_INSTRUMENTATION_INJECTION_MODE"

func findEnvVar(env []corev1.EnvVar, name string) (corev1.EnvVar, bool) {
	for _, e := range env {
		if e.Name == name {
			return e, true
		}
	}
	return corev1.EnvVar{}, false
}

func Test_apm_injectionMode_envVar_only_when_explicitly_configured(t *testing.T) {
	tests := []struct {
		name        string
		overrides   map[string]string
		wantPresent bool
		wantValue   string
	}{
		{
			name:        "default values - env var is not set",
			overrides:   map[string]string{},
			wantPresent: false,
		},
		{
			name: "explicit injectionMode - env var is set",
			overrides: map[string]string{
				"datadog.apm.instrumentation.injectionMode": "init_container",
			},
			wantPresent: true,
			wantValue:   "init_container",
		},
		{
			name: "explicit injectionMode image_volume - env var is set",
			overrides: map[string]string{
				"datadog.apm.instrumentation.injectionMode": "image_volume",
			},
			wantPresent: true,
			wantValue:   "image_volume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				// Avoid coupling this test to secret rendering behavior.
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
			}
			for k, v := range tt.overrides {
				overrides[k] = v
			}

			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.NoError(t, err, "failed to render chart")

			var deployment appsv1.Deployment
			common.Unmarshal(t, manifest, &deployment)
			require.NotEmpty(t, deployment.Spec.Template.Spec.Containers, "expected at least one container in cluster-agent deployment")

			dcaContainer := deployment.Spec.Template.Spec.Containers[0]
			envVar, found := findEnvVar(dcaContainer.Env, ddApmInjectionModeEnvVar)

			if !tt.wantPresent {
				require.False(t, found, "did not expect %s to be present", ddApmInjectionModeEnvVar)
				return
			}

			require.True(t, found, "expected %s to be present", ddApmInjectionModeEnvVar)
			require.Equal(t, tt.wantValue, envVar.Value)
		})
	}
}
