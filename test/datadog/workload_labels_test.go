package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	partOfLabelKey         = "app.kubernetes.io/part-of"
	instanceLabelKey       = "app.kubernetes.io/instance"
	agentComponentLabelKey = "agent.datadoghq.com/component"
)

func Test_workload_labels(t *testing.T) {
	tests := []struct {
		name           string
		command        common.HelmCommand
		expectedPartOf string
		expectedName   string
	}{
		{
			name: "default (escape hyphen in namespace)",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "datadog--agent-datadog",
			expectedName:   "datadog",
		},
		{
			name: "minimal",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				Namespace:   "default",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "default-datadog",
			expectedName:   "datadog",
		},
		{
			name: "escape hyphen in release name",
			command: common.HelmCommand{
				ReleaseName: "datadog-agent",
				Namespace:   "default",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "default-datadog--agent",
			expectedName:   "datadog-agent",
		},
		{
			name: "escape hyphen in both namespace and release name",
			command: common.HelmCommand{
				ReleaseName: "datadog-agent",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "datadog--agent-datadog--agent",
			expectedName:   "datadog-agent",
		},
		{
			name: "with nameOverride",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				Namespace:   "default",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"nameOverride":                 "custom",
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "default-datadog--custom",
			expectedName:   "datadog-custom",
		},
		{
			name: "with fullnameOverride",
			command: common.HelmCommand{
				ReleaseName: "ddog",
				Namespace:   "ns",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"fullnameOverride":             "superdog",
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "ns-superdog",
			expectedName:   "superdog",
		},
		{
			name: "fullnameOverride has priority over nameOverride and release",
			command: common.HelmCommand{
				ReleaseName: "datadog-agent",
				Namespace:   "default",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"nameOverride":                 "ignored-name",
					"fullnameOverride":             "expected-custom",
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "default-expected--custom", // namespace with hyphens escaped; adjust via helm render if needed
			expectedName:   "expected-custom",
		},
		{
			name: "workload labels not longer than 63 chars",
			command: common.HelmCommand{
				ReleaseName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Namespace:   "bbbbbbbbbbbbbbbbbbbbbbbbb",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "bbbbbbbbbbbbbbbbbbbbbbbbb-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa--da",
			expectedName:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-datadog",
		},
		{
			name: "workload labels not longer than 63 chars with hyphens",
			command: common.HelmCommand{
				ReleaseName: "aaaaaaaaaaaaaa-aaaaaaaaaaaaaaaaaa",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "datadog--agent-aaaaaaaaaaaaaa--aaaaaaaaaaaaaaaaaa--datadog",
			expectedName:   "aaaaaaaaaaaaaa-aaaaaaaaaaaaaaaaaa-datadog",
		},
		{
			name: "workload labels not longer than 63 chars with long namespace and release",
			command: common.HelmCommand{
				ReleaseName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Namespace:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb-aaaaaaaaaaaaaaaaaaa",
			expectedName:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaa-datadog",
		},
		{
			name: "part-of label not longer than 63 chars and trailing `--` hyphens are trimmed",
			command: common.HelmCommand{
				ReleaseName: "aaaaaaaaaaaaaaaaaaa-xxx",
				Namespace:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb-aaaaaaaaaaaaaaaaaaa",
			expectedName:   "aaaaaaaaaaaaaaaaaaa-xxx-datadog",
		},
		{
			name: "part-of label not longer than 63 chars and trailing `-` hyphen is trimmed",
			command: common.HelmCommand{
				ReleaseName: "a",
				Namespace:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml",
					"templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			expectedName:   "a-datadog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")

			manifests := strings.Split(manifest, "---")[1:]
			require.Len(t, manifests, 3)

			agent, dca, ccr := manifests[0], manifests[1], manifests[2]

			verifyDsLabels(t, agent, tt.expectedPartOf, tt.expectedName)
			verifyDcaLabels(t, dca, tt.expectedPartOf, tt.expectedName)
			verifyCcrLabels(t, ccr, tt.expectedPartOf, tt.expectedName)

		})
	}
}

func verifyDsLabels(t *testing.T, manifest string, expectedPartOf string, expectedName string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)

	labels := ds.GetLabels()
	templateLabels := ds.Spec.Template.GetLabels()

	for _, l := range []map[string]string{labels, templateLabels} {
		assert.Contains(t, l, partOfLabelKey)
		assert.Equal(t, expectedPartOf, l[partOfLabelKey])
		assert.LessOrEqual(t, len(l[partOfLabelKey]), 63)

		assert.Contains(t, l, instanceLabelKey)
		assert.Equal(t, fmt.Sprintf("%s-agent", expectedName), l[instanceLabelKey])
		assert.LessOrEqual(t, len(l[instanceLabelKey]), 63)

		assert.Contains(t, l, agentComponentLabelKey)
		assert.Equal(t, "agent", l[agentComponentLabelKey])
		assert.LessOrEqual(t, len(l[agentComponentLabelKey]), 63)
	}
}

func verifyDcaLabels(t *testing.T, manifest string, expectedPartOf string, expectedName string) {
	var dep appsv1.Deployment
	common.Unmarshal(t, manifest, &dep)

	labels := dep.GetLabels()
	templateLabels := dep.Spec.Template.GetLabels()

	for _, l := range []map[string]string{labels, templateLabels} {
		assert.Contains(t, l, partOfLabelKey)
		assert.Equal(t, expectedPartOf, l[partOfLabelKey])
		assert.LessOrEqual(t, len(l[partOfLabelKey]), 63)

		assert.Contains(t, l, instanceLabelKey)
		assert.Equal(t, fmt.Sprintf("%s-cluster-agent", expectedName), l[instanceLabelKey])
		assert.LessOrEqual(t, len(l[instanceLabelKey]), 63)

		assert.Contains(t, l, agentComponentLabelKey)
		assert.Equal(t, "cluster-agent", l[agentComponentLabelKey])
		assert.LessOrEqual(t, len(l[agentComponentLabelKey]), 63)

	}
}

func verifyCcrLabels(t *testing.T, manifest string, expectedPartOf string, expectedName string) {
	var dep appsv1.Deployment
	common.Unmarshal(t, manifest, &dep)

	labels := dep.GetLabels()
	templateLabels := dep.Spec.Template.GetLabels()

	for _, l := range []map[string]string{labels, templateLabels} {
		assert.Contains(t, l, partOfLabelKey)
		assert.Equal(t, expectedPartOf, l[partOfLabelKey])
		assert.LessOrEqual(t, len(l[partOfLabelKey]), 63)

		assert.Contains(t, l, instanceLabelKey)
		assert.Equal(t, fmt.Sprintf("%s-cluster-checks-runner", expectedName), l[instanceLabelKey])
		assert.LessOrEqual(t, len(l[instanceLabelKey]), 63)

		assert.Contains(t, l, agentComponentLabelKey)
		assert.Equal(t, "cluster-checks-runner", l[agentComponentLabelKey])
		assert.LessOrEqual(t, len(l[agentComponentLabelKey]), 63)
	}
}
