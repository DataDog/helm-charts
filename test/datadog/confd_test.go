package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestConfd(t *testing.T) {
	tests := []struct {
		name      string
		command   common.HelmCommand
		assertion func(t *testing.T, manifest string)
	}{
		{
			name: "Agent confd",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/daemonset.yaml",
					"templates/confd-configmap.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml", "../../charts/datadog/ci/confd-values.yaml"},
			},
			assertion: func(t *testing.T, manifest string) {
				var ds appsv1.DaemonSet
				common.Unmarshal(t, manifest, &ds)

				// Check that the confd volume is present
				var confdVolume *corev1.Volume = nil
				for _, vol := range ds.Spec.Template.Spec.Volumes {
					if vol.Name == "confd" {
						confdVolume = &vol
						break
					}
				}
				require.NotNil(t, confdVolume, "confd volume not found in daemonset")
				require.NotNil(t, confdVolume.ConfigMap, "confd volume should be a ConfigMap volume")
				require.Equal(t, "datadog-confd", confdVolume.ConfigMap.Name, "unexpected ConfigMap name for confd volume")

				// Check that the container has the confd volume mount
				var initConfig corev1.Container
				for _, init := range ds.Spec.Template.Spec.InitContainers {
					if init.Name == "init-config" {
						initConfig = init
						break
					}
				}
				var confdMount *corev1.VolumeMount = nil
				for _, mount := range initConfig.VolumeMounts {
					if mount.Name == "confd" {
						confdMount = &mount
						break
					}
				}
				require.NotNil(t, confdMount, "confd volume mount not found in container")
				require.Equal(t, "/conf.d", confdMount.MountPath, "unexpected mount path for confd volume")
			},
		},
		{
			name: "Cluster agent confd and advancedConfd",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				Namespace:   "datadog-agent",
				ChartPath:   "../../charts/datadog",
				ShowOnly: []string{
					"templates/cluster-agent-deployment.yaml",
					"templates/cluster-agent-confd-configmap.yaml",
				},
				Values: []string{"../../charts/datadog/values.yaml", "../../charts/datadog/ci/cluster-agent-advanced-confd-values.yaml"},
			},
			assertion: func(t *testing.T, manifest string) {
				var deployment appsv1.Deployment
				common.Unmarshal(t, manifest, &deployment)

				// Check that the confd volume is present
				var confdVolume *corev1.Volume = nil
				for _, vol := range deployment.Spec.Template.Spec.Volumes {
					if vol.Name == "confd" {
						confdVolume = &vol
						break
					}
				}
				require.NotNil(t, confdVolume, "confd volume not found in deployment")
				require.NotNil(t, confdVolume.ConfigMap, "confd volume should be a ConfigMap volume")
				require.Equal(t, "datadog-cluster-agent-confd", confdVolume.ConfigMap.Name, "unexpected ConfigMap name for confd volume")

				// Check that the volume has the expected items
				expectedItems := map[string]string{
					"redisdb.yaml":           "redisdb.yaml",
					"orchestrator.d--1.yaml": "orchestrator.d/1.yaml",
					"orchestrator.d--2.yaml": "orchestrator.d/2.yaml",
				}

				require.Len(t, confdVolume.ConfigMap.Items, len(expectedItems), "unexpected number of items in confd volume")

				for _, item := range confdVolume.ConfigMap.Items {
					expectedPath, ok := expectedItems[item.Key]
					require.True(t, ok, "unexpected key in confd volume: %s", item.Key)
					require.Equal(t, expectedPath, item.Path, "unexpected path for key %s", item.Key)
				}

				// Check that the container has the confd volume mount
				container := deployment.Spec.Template.Spec.Containers[0]
				var confdMount *corev1.VolumeMount = nil
				for _, mount := range container.VolumeMounts {
					if mount.Name == "confd" {
						confdMount = &mount
						break
					}
				}
				require.NotNil(t, confdMount, "confd volume mount not found in container")
				require.Equal(t, "/conf.d", confdMount.MountPath, "unexpected mount path for confd volume")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			require.NoError(t, err, "failed to render chart")

			tt.assertion(t, manifest)
		})
	}
}
