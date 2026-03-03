package utils

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func VerifyBaseline[T any](t *testing.T, baselineManifestPath, manifest string, baseline, actual T) {
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)

	// Exclude "helm.sh/chart" label from comparison to avoid
	// updating baselines on every unrelated chart changes.
	ops := make(cmp.Options, 0)
	ops = append(ops, cmpopts.IgnoreMapEntries(func(k, v string) bool {
		return k == "helm.sh/chart"
	}))

	assert.True(t, cmp.Equal(baseline, actual, ops), cmp.Diff(baseline, actual))
}
