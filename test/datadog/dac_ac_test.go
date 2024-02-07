package datadog

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_admissionControllerConfig(t *testing.T) {
	tests := []struct {
		name                 string
		command              common.HelmCommand
		baselineManifestPath string
		assertions           func(t *testing.T, baselineManifestPath, manifest string)
	}{
		{
			name: "DCA Deployment default",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"clusterAgent.admissionController.enabled":                       "true",
					"clusterAgent.admissionController.agentSidecarInjection.enabled": "true",
				},
			},
			baselineManifestPath: "./baseline/cluster-agent-deployment_default.yaml",
			assertions:           verifyDeploymentACConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			t.Log("update baselines", common.UpdateBaselines)
			if common.UpdateBaselines {
				common.WriteToFile(t, tt.baselineManifestPath, manifest)
			}
			tt.assertions(t, tt.baselineManifestPath, manifest)
		})
	}
}

type Selector struct {
	ObjectSelector    metav1.LabelSelector `yaml:"objectSelector"`
	NamespaceSelector metav1.LabelSelector `yaml:"namespaceSelector"`
}

func verifyDeploymentACConfig(t *testing.T, baselineManifestPath, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	dcaContainer := deployment.Spec.Template.Spec.Containers[0]
	var selectorsAsString string
	for _, envVar := range dcaContainer.Env {
		if "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_SELECTORS" == envVar.Name {
			selectorsAsString = envVar.Value
		}
	}

	var selectors []Selector
	json.Unmarshal([]byte(selectorsAsString), &selectors)
	t.Log("print", "selector struct", selectors, "selector string", selectorsAsString)
}
