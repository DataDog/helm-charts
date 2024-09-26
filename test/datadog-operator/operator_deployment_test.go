package datadog_operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

// This test will produce two renderings for two versions of DatadogAgent.
// Will convert v1alpha1 to v2alpha1 and compare to rendered v2alpha1.
//
// Rendering is done by Terratest, for below inputs it will run helm command:
//
// helm template --set useV2alpha1=false \
//	             --show-only "templates/datadogagent.yaml \
//	             -f ../k8s/datadog-agent-with-operator/values/staging.yaml \
//	             -f ../charts/.common_lint_values.yaml \
//	             datadog-operator "[path to the charts folder]/datadog-agent-with-operator"

const (
	SkipTest = false
)

func Test_operator_chart(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertions func(t *testing.T, manifest string)
		skipTest   bool
	}{
		{
			name: "Verify Operator 1.0 deployment",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   map[string]string{},
			},
			assertions: verifyDeployment,
			skipTest:   SkipTest,
		},
		{
			name: "Rendering all does not fail",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   map[string]string{},
			},
			assertions: verifyAll,
			skipTest:   SkipTest,
		},
		{
			name: "livenessProbe is correctly configured",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   map[string]string{},
			},
			assertions: verifyLivenessProbe,
			skipTest:   SkipTest,
		},
		{
			name: "livenessProbe is correctly overriden",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"livenessProbe.timeoutSeconds":   "20",
					"livenessProbe.periodSeconds":    "20",
					"livenessProbe.failureThreshold": "3",
				},
			},
			assertions: verifyLivenessProbeOverride,
			skipTest:   SkipTest,
		},
	}

	for _, tt := range tests {
		if tt.skipTest {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			manifest, err1 := common.RenderChart(t, tt.command)
			assert.Nil(t, err1, "can't render template", "command", tt.command)
			tt.assertions(t, manifest)
		})
	}
}

func verifyDeployment(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Containers))
	operatorContainer := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, v1.PullPolicy("IfNotPresent"), operatorContainer.ImagePullPolicy)
	assert.Equal(t, "gcr.io/datadoghq/operator:1.8.0", operatorContainer.Image)
	assert.NotContains(t, operatorContainer.Args, "-webhookEnabled=false")
	assert.NotContains(t, operatorContainer.Args, "-webhookEnabled=true")
}

func verifyAll(t *testing.T, manifest string) {
	assert.True(t, manifest != "")
}

func verifyLivenessProbe(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Containers))
	operatorContainer := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "/healthz/", operatorContainer.LivenessProbe.HTTPGet.Path)
}

func verifyLivenessProbeOverride(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Containers))
	operatorContainer := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "/healthz/", operatorContainer.LivenessProbe.HTTPGet.Path)
	assert.Equal(t, int32(20), operatorContainer.LivenessProbe.PeriodSeconds)
	assert.Equal(t, int32(20), operatorContainer.LivenessProbe.TimeoutSeconds)
	assert.Equal(t, int32(3), operatorContainer.LivenessProbe.FailureThreshold)
}
