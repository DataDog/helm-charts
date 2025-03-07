package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
)

func TestAPMConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		command common.HelmCommand
		isValid bool
	}{
		{
			name: "valid enabled configuration",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/valid_enabled.yaml",
				},
			},
			isValid: true,
		},
		{
			name: "valid target configuration",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/valid_targets.yaml",
				},
			},
			isValid: true,
		},
		{
			name: "valid namespace configuration",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/valid_namespace.yaml",
				},
			},
			isValid: true,
		},
		{
			name: "both namespaces and targets",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/namespaces_and_targets.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "both libversions and targets",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/libversions_and_targets.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "both enabled and disabled namespaces",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/enabled_and_disabled_namespaces.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "both matchLabels and matchNames for namespace selector",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/namespace_labels_and_names.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "both matchExpressions and matchNames for namespace selector",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/namespace_exprs_and_names.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "extraneous instrumentation key",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/extra_instrumentation_key.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "extraneous target key",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/extra_target_key.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "extraneous pod selector key",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/extra_podselector_key.yaml",
				},
			},
			isValid: false,
		},
		{
			name: "extraneous namespace selector key",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values: []string{
					"../../charts/datadog/values.yaml",
					"values/instrumentation/extra_namespaceselector_key.yaml",
				},
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := common.RenderChart(t, tt.command)
			if tt.isValid {
				assert.Nil(t, err, "expected no error, got %v", err)
			} else {
				assert.NotNil(t, err, "expected error, got nil")
			}
		})
	}
}
