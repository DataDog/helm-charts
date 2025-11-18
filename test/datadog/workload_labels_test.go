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
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
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
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
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
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
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
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
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
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
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
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
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
			name: "labels not longer than 63 chars",
			command: common.HelmCommand{
				ReleaseName: "supersuperdupercalifragilisticexpialidocious",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "datadog--agent-supersuperdupercalifragilisticexpialidoc",
			expectedName:   "supersuperdupercalifragilisticexpialidoc",
		},
		{
			name: "labels not longer than 63 chars with hyphens",
			command: common.HelmCommand{
				ReleaseName: "super-superdupercalifragilisticexpialidocious",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/cluster-agent-deployment.yaml", "templates/agent-clusterchecks-deployment.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"clusterChecksRunner.enabled":  "true",
				},
			},
			expectedPartOf: "datadog--agent-super--superdupercalifragilisticexpialido",
			expectedName:   "super-superdupercalifragilisticexpialido",
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

		assert.Contains(t, l, instanceLabelKey)
		assert.Equal(t, fmt.Sprintf("%s-agent", expectedName), l[instanceLabelKey])

		assert.Contains(t, l, agentComponentLabelKey)
		assert.Equal(t, "agent", l[agentComponentLabelKey])
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

		assert.Contains(t, l, instanceLabelKey)
		assert.Equal(t, fmt.Sprintf("%s-cluster-agent", expectedName), l[instanceLabelKey])

		assert.Contains(t, l, agentComponentLabelKey)
		assert.Equal(t, "cluster-agent", l[agentComponentLabelKey])
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

		assert.Contains(t, l, instanceLabelKey)
		assert.Equal(t, fmt.Sprintf("%s-cluster-checks-runner", expectedName), l[instanceLabelKey])

		assert.Contains(t, l, agentComponentLabelKey)
		assert.Equal(t, "cluster-checks-runner", l[agentComponentLabelKey])
	}
}
