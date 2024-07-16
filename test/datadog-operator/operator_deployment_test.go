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
			name: "Verify Operator 1.0 cert secret name",
			command: common.HelmCommand{
				ReleaseName: "random-string-as-release-name",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.useCertManager":            "true",
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "true",
				},
			},
			assertions: verifyDeploymentCertSecretName,
			skipTest:   SkipTest,
		},
		{
			name: "Verify Operator 1.0 conversionWebhook.enabled=true",
			command: common.HelmCommand{
				ReleaseName: "random-string-as-release-name",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "true",
				},
			},
			assertions: verifyConversionWebhookEnabledTrue,
			skipTest:   SkipTest,
		},
		{
			name: "Verify Operator 1.0 conversionWebhook.enabled=false",
			command: common.HelmCommand{
				ReleaseName: "random-string-as-release-name",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "false",
				},
			},
			assertions: verifyConversionWebhookEnabledFalse,
			skipTest:   SkipTest,
		},
		{
			name: "Verify Operator 1.0 conversionWebhook.enabled default",
			command: common.HelmCommand{
				ReleaseName: "random-string-as-release-name",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
			},
			assertions: verifyConversionWebhookEnabledFalse,
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
	assert.Equal(t, "gcr.io/datadoghq/operator:1.7.0", operatorContainer.Image)
	assert.NotContains(t, operatorContainer.Args, "-webhookEnabled=false")
}

func verifyDeploymentCertSecretName(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)

	var mode = int32(420)
	assert.Contains(t, deployment.Spec.Template.Spec.Volumes, v1.Volume{
		Name: "cert",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				DefaultMode: &mode,
				SecretName:  "random-string-as-release-name-webhook-server-cert",
			},
		},
	})
}

func verifyConversionWebhookEnabledTrue(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	operatorContainer := deployment.Spec.Template.Spec.Containers[0]
	assert.NotContains(t, operatorContainer.Args, "-webhookEnabled=true")
}

func verifyConversionWebhookEnabledFalse(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	operatorContainer := deployment.Spec.Template.Spec.Containers[0]
	assert.NotContains(t, operatorContainer.Args, "-webhookEnabled=false")
}

func verifyAll(t *testing.T, manifest string) {
	assert.True(t, manifest != "")
}
