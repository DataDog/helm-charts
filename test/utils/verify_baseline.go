package utils

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

// Normalizer mutates a loaded manifest in place before it's compared, e.g. to
// strip out fields that are expected to vary independently of chart
// structure (see stripImageTag in test/datadog-operator/baseline_test.go).
type Normalizer[T any] func(*T)

func VerifyBaseline[T any](t *testing.T, baselineManifestPath, manifest string, baseline, actual T, normalizers ...Normalizer[T]) {
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)

	for _, normalize := range normalizers {
		normalize(&actual)
		normalize(&baseline)
	}

	// Exclude "helm.sh/chart" label from comparison to avoid
	// updating baselines on every unrelated chart changes.
	ops := cmp.Options{
		cmpopts.IgnoreMapEntries(func(k, v string) bool {
			return k == "helm.sh/chart"
		}),
	}

	assert.True(t, cmp.Equal(baseline, actual, ops), cmp.Diff(baseline, actual, ops))
}
