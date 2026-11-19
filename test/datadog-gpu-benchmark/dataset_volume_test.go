package datadog_gpu_benchmark

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/DataDog/helm-charts/test/common"
)

const chartPath = "../../charts/datadog-gpu-benchmark"

// The chart's required values, and the namespace Datadog creates benchmark Jobs in.
var required = map[string]string{
	"clusterAgent.namespace":      "datadog",
	"clusterAgent.serviceAccount": "datadog-cluster-agent",
}

func render(t *testing.T, overrides map[string]string, overridesJSON map[string]string, showOnly ...string) (string, error) {
	t.Helper()
	set := map[string]string{}
	for k, v := range required {
		set[k] = v
	}
	for k, v := range overrides {
		set[k] = v
	}
	return common.RenderChart(t, common.HelmCommand{
		ReleaseName:   "datadog-gpu-benchmark",
		Namespace:     "datadog-gpu-benchmark",
		ChartPath:     chartPath,
		ShowOnly:      showOnly,
		Overrides:     set,
		OverridesJson: overridesJSON,
	})
}

func renderClaim(t *testing.T, overrides map[string]string, overridesJSON map[string]string) corev1.PersistentVolumeClaim {
	t.Helper()
	manifest, err := render(t, overrides, overridesJSON, "templates/dataset-volume.yaml")
	require.NoError(t, err)
	var claim corev1.PersistentVolumeClaim
	common.Unmarshal(t, manifest, &claim)
	return claim
}

func TestDatasetVolumeDisabledByDefault(t *testing.T) {
	_, err := render(t, nil, nil, "templates/dataset-volume.yaml")
	// helm template fails when --show-only matches no rendered manifest.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not find template")
}

func TestDatasetVolume(t *testing.T) {
	claim := renderClaim(t, map[string]string{"datasetVolume.enabled": "true"}, nil)

	assert.Equal(t, "datadog-gpu-benchmark-datasets", claim.Name)
	assert.Equal(t, "datadog-gpu-benchmark", claim.Namespace)
	assert.Equal(t, "keep", claim.Annotations["helm.sh/resource-policy"], "datasets outlive the release")
	assert.Equal(t, []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}, claim.Spec.AccessModes)
	assert.Equal(t, resource.MustParse("50Gi"), claim.Spec.Resources.Requests[corev1.ResourceStorage])
	assert.Nil(t, claim.Spec.StorageClassName, "unset uses the cluster's default StorageClass")
	assert.Empty(t, claim.Spec.VolumeName)
}

func TestDatasetVolumeStorageClass(t *testing.T) {
	claim := renderClaim(t, map[string]string{
		"datasetVolume.enabled":          "true",
		"datasetVolume.storageClassName": "efs",
		"datasetVolume.size":             "100Gi",
	}, map[string]string{"datasetVolume.accessModes": `["ReadWriteOnce"]`})

	require.NotNil(t, claim.Spec.StorageClassName)
	assert.Equal(t, "efs", *claim.Spec.StorageClassName)
	assert.Equal(t, resource.MustParse("100Gi"), claim.Spec.Resources.Requests[corev1.ResourceStorage])
	assert.Equal(t, []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}, claim.Spec.AccessModes)
}

// A PersistentVolume a bucket CSI driver serves often has no StorageClass: the claim must ask
// for none explicitly, or it gets the cluster's default and never binds.
func TestDatasetVolumeClasslessPersistentVolume(t *testing.T) {
	claim := renderClaim(t, map[string]string{
		"datasetVolume.enabled":    "true",
		"datasetVolume.volumeName": "s3-datasets",
	}, map[string]string{"datasetVolume.storageClassName": `""`})

	assert.Equal(t, "s3-datasets", claim.Spec.VolumeName)
	require.NotNil(t, claim.Spec.StorageClassName, "rendered, not omitted")
	assert.Equal(t, "", *claim.Spec.StorageClassName)
}

func TestDatasetVolumeRejectsInvalidAccessModes(t *testing.T) {
	for name, modes := range map[string]string{
		"empty":   `[]`,
		"unknown": `["Bogus"]`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := render(t, map[string]string{"datasetVolume.enabled": "true"},
				map[string]string{"datasetVolume.accessModes": modes}, "templates/dataset-volume.yaml")
			assert.Error(t, err)
		})
	}
}
