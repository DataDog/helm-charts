package datadog

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

// V1 structs are for the current scope
type SelectorV1 struct {
	ObjectSelector    metav1.LabelSelector `json:"objectSelector"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector"`
}

type ProfileV1 struct {
	EnvVars              []corev1.EnvVar             `json:"env,omitempty"`
	ResourceRequirements corev1.ResourceRequirements `json:"resources,omitempty"`
}

// V2 structs are for one possibility of extending V1
type SelectorV2 struct {
	Name string `yaml:"name"`

	ObjectSelector    metav1.LabelSelector `json:"objectSelector"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector"`
}

type ProfileV2 struct {
	Name           string   `yaml:"name"`
	Default        bool     `yaml:"default"`
	basedOnDefault bool     `yaml:"basedOnDefault"`
	Selectors      []string `yaml:selectors`

	EnvVars              []corev1.EnvVar             `json:"env,omitempty"`
	ResourceRequirements corev1.ResourceRequirements `json:"resources,omitempty"`
}

func verifyDeploymentACConfig(t *testing.T, baselineManifestPath, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	dcaContainer := deployment.Spec.Template.Spec.Containers[0]
	var selectorsAsString, profilesAsString string
	for _, envVar := range dcaContainer.Env {
		if "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_SELECTORS" == envVar.Name {
			selectorsAsString = envVar.Value
		}
		if "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_PROFILES" == envVar.Name {
			profilesAsString = envVar.Value
		}
	}

	// Unmarshal into v1 struct
	var selectorsV1 []SelectorV1
	err := json.Unmarshal([]byte(selectorsAsString), &selectorsV1)
	t.Log("print", "selector struct", selectorsV1, "selector string", selectorsAsString, "error", err)

	var profilesV1 []ProfileV1
	err = json.Unmarshal([]byte(profilesAsString), &profilesV1)
	t.Log("print", "profile struct", profilesV1, "profile string", profilesAsString, "error", err)

	// Unmarshal into v2 structs
	var selectorsV2 []SelectorV2
	err = json.Unmarshal([]byte(selectorsAsString), &selectorsV2)
	t.Log("print", "selector struct", selectorsV2, "selector string", selectorsAsString, "error", err)

	var profilesV2 []ProfileV2
	err = json.Unmarshal([]byte(profilesAsString), &profilesV2)
	t.Log("print", "profile struct", profilesV2, "profile string", profilesAsString, "error", err)

	assert.Equal(t, 2, len(selectorsV1))
	assert.Equal(t, 2, len(selectorsV2))
	assert.Equal(t, selectorsV1[0].NamespaceSelector, selectorsV2[0].NamespaceSelector)
	assert.Equal(t, selectorsV1[1].NamespaceSelector, selectorsV2[1].NamespaceSelector)
	assert.Equal(t, selectorsV1[0].ObjectSelector, selectorsV2[0].ObjectSelector)
	assert.Equal(t, selectorsV1[1].ObjectSelector, selectorsV2[1].ObjectSelector)

	assert.Equal(t, 2, len(profilesV1))
	assert.Equal(t, 2, len(profilesV2))
	assert.Equal(t, profilesV1[0].EnvVars, profilesV1[0].EnvVars)
	assert.Equal(t, profilesV2[1].EnvVars, profilesV2[1].EnvVars)
	assert.Equal(t, profilesV1[0].ResourceRequirements, profilesV1[0].ResourceRequirements)
	assert.Equal(t, profilesV2[1].ResourceRequirements, profilesV2[1].ResourceRequirements)

	assert.Equal(t, "fargate-profile1", selectorsV2[0].Name)
	assert.Equal(t, true, profilesV2[0].Default)
}
