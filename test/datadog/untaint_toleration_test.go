package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

// untaintToleration is the toleration the chart must inject so the node Agent
// can schedule on nodes carrying the Datadog Operator untaint controller's
// startup taint agent.datadoghq.com/not-ready=presence:NoSchedule.
var untaintToleration = corev1.Toleration{
	Key:      "agent.datadoghq.com/not-ready",
	Operator: corev1.TolerationOpEqual,
	Value:    "presence",
	Effect:   corev1.TaintEffectNoSchedule,
}

func Test_untaintToleration(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertions func(t *testing.T, manifest string)
	}{
		{
			name: "disabled by default -- toleration absent",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
				},
			},
			assertions: verifyUntaintTolerationAbsent,
		},
		{
			name: "enabled -- toleration injected",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":     "datadog-secret",
					"datadog.appKeyExistingSecret":     "datadog-secret",
					"agents.untaintToleration.enabled": "true",
				},
			},
			assertions: verifyUntaintTolerationPresent,
		},
		{
			name: "enabled alongside user tolerations -- both present",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":     "datadog-secret",
					"datadog.appKeyExistingSecret":     "datadog-secret",
					"agents.untaintToleration.enabled": "true",
					"agents.tolerations[0].key":        "dedicated",
					"agents.tolerations[0].operator":   "Exists",
					"agents.tolerations[0].effect":     "NoSchedule",
				},
			},
			assertions: verifyUntaintTolerationWithUserToleration,
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

func hasToleration(tolerations []corev1.Toleration, want corev1.Toleration) bool {
	for _, tol := range tolerations {
		if tol == want {
			return true
		}
	}
	return false
}

func verifyUntaintTolerationAbsent(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)
	assert.False(t, hasToleration(ds.Spec.Template.Spec.Tolerations, untaintToleration),
		"untaint toleration must not be present when agents.untaintToleration.enabled is false")
}

func verifyUntaintTolerationPresent(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)
	assert.True(t, hasToleration(ds.Spec.Template.Spec.Tolerations, untaintToleration),
		"untaint toleration must be injected when agents.untaintToleration.enabled is true")
}

func verifyUntaintTolerationWithUserToleration(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)
	tolerations := ds.Spec.Template.Spec.Tolerations
	assert.True(t, hasToleration(tolerations, untaintToleration),
		"untaint toleration must be injected")
	assert.True(t, hasToleration(tolerations, corev1.Toleration{
		Key:      "dedicated",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}), "user-provided agents.tolerations must be preserved alongside the untaint toleration")
}
