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

const (
	DDSidecarEnabled             = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_ENABLED"
	DDSidecarClusterAgentEnabled = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_CLUSTER_AGENT_ENABLED"
	DDSidecarProvider            = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_PROVIDER"
	DDSidecarRegistry            = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_CONTAINER_REGISTRY"
	DDSidecarImageName           = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_IMAGE_NAME"
	DDSidecarImageTag            = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_IMAGE_TAG"
	DDSidecarSelectors           = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_SELECTORS"
	DDSidecarProfiles            = "DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_PROFILES"
)

func Test_admissionControllerConfig(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertions func(t *testing.T, manifest string)
	}{
		{
			name: "AC sidecar injection, minimal Fargate config",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values: []string{"../../charts/datadog/values.yaml",
					"./manifests/dca_AC_sidecar_fargateMinimal.yaml"},
				Overrides: map[string]string{
					// "clusterAgent.admissionController.enabled":                       "true",
					// "clusterAgent.admissionController.agentSidecarInjection.enabled": "true",
				},
			},
			assertions: verifyDeploymentFargateMinimal,
		},
		{
			name: "AC sidecar injection, advanced config",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values: []string{"../../charts/datadog/values.yaml",
					"./manifests/dca_AC_sidecar_advanced.yaml"},
				Overrides: map[string]string{},
			},
			assertions: verifyDeploymentAdvancedConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			tt.assertions(t, manifest)
		})
	}
}

// V1 structs are for the current scope
type Selector struct {
	ObjectSelector    metav1.LabelSelector `json:"objectSelector,omitempty"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

type ProfileOverride struct {
	EnvVars              []corev1.EnvVar             `json:"env,omitempty"`
	ResourceRequirements corev1.ResourceRequirements `json:"resources,omitempty"`
}

func verifyDeploymentFargateMinimal(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	dcaContainer := deployment.Spec.Template.Spec.Containers[0]

	acConfigEnv := selectEnvVars(dcaContainer.Env)

	assert.Equal(t, "true", acConfigEnv[DDSidecarEnabled])
	assert.Equal(t, "true", acConfigEnv[DDSidecarClusterAgentEnabled])
	assert.Equal(t, "fargate", acConfigEnv[DDSidecarProvider])
	// Default will be set by DCA
	assert.Empty(t, acConfigEnv[DDSidecarRegistry])
	assert.Equal(t, "agent", acConfigEnv[DDSidecarImageName])
	assert.Equal(t, "7.51.0", acConfigEnv[DDSidecarImageTag])
	assert.Empty(t, acConfigEnv[DDSidecarSelectors])
	assert.Empty(t, acConfigEnv[DDSidecarProfiles])
}

func verifyDeploymentAdvancedConfig(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	dcaContainer := deployment.Spec.Template.Spec.Containers[0]

	acConfigEnv := selectEnvVars(dcaContainer.Env)

	assert.Equal(t, "true", acConfigEnv[DDSidecarEnabled])
	assert.Equal(t, "false", acConfigEnv[DDSidecarClusterAgentEnabled])
	assert.Empty(t, acConfigEnv[DDSidecarProvider])
	assert.Equal(t, "gcr.io/datadoghq", acConfigEnv[DDSidecarRegistry])
	assert.Equal(t, "agent", acConfigEnv[DDSidecarImageName])
	assert.Equal(t, "7.52.0", acConfigEnv[DDSidecarImageTag])
	assert.NotEmpty(t, acConfigEnv[DDSidecarSelectors])
	assert.NotEmpty(t, acConfigEnv[DDSidecarProfiles])

	selectorsAsString := acConfigEnv[DDSidecarSelectors]
	profilesAsString := acConfigEnv[DDSidecarProfiles]

	var selectors []Selector
	err := json.Unmarshal([]byte(selectorsAsString), &selectors)
	assert.Nil(t, err)
	selector := selectors[0]
	assert.Equal(t, "nodeless", selector.ObjectSelector.MatchLabels["runsOn"])
	assert.Equal(t, "nginx", selector.ObjectSelector.MatchLabels["app"])
	assert.Equal(t, "true", selector.NamespaceSelector.MatchLabels["agentSidecars"])

	var profiles []ProfileOverride
	err = json.Unmarshal([]byte(profilesAsString), &profiles)
	assert.Nil(t, err)
	profile := profiles[0]
	assert.Equal(t, "DD_ORCHESTRATOR_EXPLORER_ENABLED", profile.EnvVars[0].Name)
	assert.Equal(t, "false", profile.EnvVars[0].Value)
	assert.Equal(t, "DD_TAGS", profile.EnvVars[1].Name)
	// Agent expects space-separated pairs
	assert.Equal(t, "key1:value1 key2:value2", profile.EnvVars[1].Value)
	assert.Equal(t, "1", profile.ResourceRequirements.Requests.Cpu().String())
	assert.Equal(t, "512Mi", profile.ResourceRequirements.Requests.Memory().String())
	assert.Equal(t, "2", profile.ResourceRequirements.Limits.Cpu().String())
	assert.Equal(t, "1Gi", profile.ResourceRequirements.Limits.Memory().String())
}

func selectEnvVars(envVars []corev1.EnvVar) map[string]string {
	acConfoigNames := []string{
		DDSidecarEnabled,
		DDSidecarClusterAgentEnabled,
		DDSidecarProvider,
		DDSidecarRegistry,
		DDSidecarImageName,
		DDSidecarImageTag,
		DDSidecarSelectors,
		DDSidecarProfiles,
	}

	selection := map[string]string{}

	for _, envVar := range envVars {
		for _, name := range acConfoigNames {
			if envVar.Name == name {
				selection[name] = envVar.Value
			}
		}
	}
	return selection
}
