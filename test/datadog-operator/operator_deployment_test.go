package datadog_operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

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
		{
			name: "Watch namespaces correctly set",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"watchNamespaces":        "{common1,common2}",
					"watchNamespacesAgent":   "{dda-ns}",
					"watchNamespacesMonitor": "{monitor-ns}",
					"watchNamespacesSLO":     "{}",
				},
			},
			assertions: verifyWatchNamespaces,
			skipTest:   SkipTest,
		},
		{
			name: "registryMigration auto: ASIA, EU, and DEFAULT overrides are set",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   map[string]string{},
			},
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				env := deployment.Spec.Template.Spec.Containers[0].Env
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_ASIA"), "ASIA should be set")
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_EU"), "EU should be set")
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_DEFAULT"), "DEFAULT should be set")
				assert.Nil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_AZURE"), "AZURE should not be set")
			},
			skipTest: SkipTest,
		},
		{
			name: "registryMigration disabled: no overrides set",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"registryMigrationMode": "",
				},
			},
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				env := deployment.Spec.Template.Spec.Containers[0].Env
				assert.Nil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_ASIA"), "ASIA should not be set")
				assert.Nil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_DEFAULT"), "DEFAULT should not be set")
				assert.Nil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_EU"), "EU should not be set")
				assert.Nil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_AZURE"), "AZURE should not be set")
			},
			skipTest: SkipTest,
		},
		{
			name: "registryMigration all: all overrides set",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"registryMigrationMode": "all",
				},
			},
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				env := deployment.Spec.Template.Spec.Containers[0].Env
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_ASIA"), "ASIA should be set")
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_DEFAULT"), "DEFAULT should be set")
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_EU"), "EU should be set")
				assert.NotNil(t, FindEnvVarByName(env, "DD_REGISTRY_OVERRIDE_AZURE"), "AZURE should be set")
			},
			skipTest: SkipTest,
		},
		{
			name: "Operator image tag with digest",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"image.tag":   "1.18.0@sha256:0000",
					"toolVersion": "unknown",
				},
			},
			skipTest: SkipTest,
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Containers))
				operatorContainer := deployment.Spec.Template.Spec.Containers[0]
				assert.Equal(t, "registry.datadoghq.com/operator:1.18.0@sha256:0000", operatorContainer.Image)
				installToolEnv := FindEnvVarByName(operatorContainer.Env, "DD_TOOL_VERSION")
				assert.Equal(t, "unknown", installToolEnv.Value)
			},
		},
		{
			name: "untaintController disabled by default",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   map[string]string{},
			},
			skipTest: SkipTest,
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				operatorContainer := deployment.Spec.Template.Spec.Containers[0]
				assert.Contains(t, operatorContainer.Args, "-untaintControllerEnabled=false")
				assert.NotContains(t, operatorContainer.Args, "-untaintControllerEnabled=true")
				// waitForCSIDriver flag and tuning env vars only render when the controller is enabled.
				assert.NotContains(t, operatorContainer.Args, "-untaintControllerWaitForCSIDriver=false")
				assert.Nil(t, FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_TIMEOUT"))
				assert.Nil(t, FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_EVENTS_ENABLED"))
			},
		},
		{
			name: "untaintController enabled sets flags",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"untaintController.enabled": "true",
				},
			},
			skipTest: SkipTest,
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				operatorContainer := deployment.Spec.Template.Spec.Containers[0]
				assert.Contains(t, operatorContainer.Args, "-untaintControllerEnabled=true")
				assert.Contains(t, operatorContainer.Args, "-untaintControllerWaitForCSIDriver=false")
				// Tuning env vars are omitted (operator defaults apply) unless explicitly set.
				assert.Nil(t, FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_TIMEOUT"))
				assert.Nil(t, FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_EVENTS_ENABLED"))
			},
		},
		{
			name: "untaintController full configuration",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"untaintController.enabled":           "true",
					"untaintController.waitForCSIDriver":  "true",
					"untaintController.timeout":           "2m",
					"untaintController.schedulingTimeout": "3m",
					"untaintController.timeoutPolicy":     "keep",
					"untaintController.eventsEnabled":     "true",
				},
			},
			skipTest: SkipTest,
			assertions: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)
				operatorContainer := deployment.Spec.Template.Spec.Containers[0]
				assert.Contains(t, operatorContainer.Args, "-untaintControllerEnabled=true")
				assert.Contains(t, operatorContainer.Args, "-untaintControllerWaitForCSIDriver=true")

				timeout := FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_TIMEOUT")
				assert.NotNil(t, timeout)
				assert.Equal(t, "2m", timeout.Value)
				schedulingTimeout := FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_SCHEDULING_TIMEOUT")
				assert.NotNil(t, schedulingTimeout)
				assert.Equal(t, "3m", schedulingTimeout.Value)
				timeoutPolicy := FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_TIMEOUT_POLICY")
				assert.NotNil(t, timeoutPolicy)
				assert.Equal(t, "keep", timeoutPolicy.Value)
				eventsEnabled := FindEnvVarByName(operatorContainer.Env, "DD_UNTAINT_CONTROLLER_EVENTS_ENABLED")
				assert.NotNil(t, eventsEnabled)
				assert.Equal(t, "true", eventsEnabled.Value)
			},
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
	assert.Equal(t, "registry.datadoghq.com/operator:1.29.0-rc.1", operatorContainer.Image)
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

func verifyWatchNamespaces(t *testing.T, manifest string) {
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Containers))
	operatorContainer := deployment.Spec.Template.Spec.Containers[0]
	watchNsEnv := FindEnvVarByName(operatorContainer.Env, "WATCH_NAMESPACE")
	agentNsEnv := FindEnvVarByName(operatorContainer.Env, "DD_AGENT_WATCH_NAMESPACE")
	monitorNsEnv := FindEnvVarByName(operatorContainer.Env, "DD_MONITOR_WATCH_NAMESPACE")
	sloNsEnv := FindEnvVarByName(operatorContainer.Env, "DD_SLO_WATCH_NAMESPACE")
	dapNsEnv := FindEnvVarByName(operatorContainer.Env, "DD_AGENT_PROFILE_WATCH_NAMESPACE")

	assert.Equal(t, "common1,common2", watchNsEnv.Value)
	assert.Equal(t, "dda-ns", agentNsEnv.Value)
	assert.Equal(t, "monitor-ns", monitorNsEnv.Value)
	assert.Equal(t, "", sloNsEnv.Value)
	assert.Nil(t, dapNsEnv)
}

func Test_operator_untaint_controller_rbac(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		wantPatch bool
	}{
		{
			name:      "untaintController disabled -- no nodes patch rule",
			overrides: map[string]string{},
			wantPatch: false,
		},
		{
			name:      "untaintController enabled -- nodes patch rule present",
			overrides: map[string]string{"untaintController.enabled": "true"},
			wantPatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/clusterrole.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   tt.overrides,
			})
			assert.Nil(t, err, "couldn't render template")

			var clusterRole rbacv1.ClusterRole
			common.Unmarshal(t, manifest, &clusterRole)
			assert.Equal(t, tt.wantPatch, hasNodesPatchRule(clusterRole.Rules),
				"unexpected presence of nodes patch rule")
		})
	}
}

func hasNodesPatchRule(rules []rbacv1.PolicyRule) bool {
	for _, rule := range rules {
		if !containsString(rule.APIGroups, "") || !containsString(rule.Resources, "nodes") {
			continue
		}
		if len(rule.Verbs) == 1 && rule.Verbs[0] == "patch" {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func FindEnvVarByName(envs []v1.EnvVar, name string) *v1.EnvVar {
	for i, env := range envs {
		if env.Name == name {
			return &envs[i]
		}
	}
	return nil
}
