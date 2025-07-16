package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
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
