package datadog

import (
	"strconv"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
)

func Test_clusterAgentPDB(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		verify    func(t *testing.T, spec map[string]string)
	}{
		{
			name: "deprecated clusterAgent.createPodDisruptionBudget true",
			overrides: map[string]string{
				"clusterAgent.createPodDisruptionBudget": "true",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMax := spec["maxUnavailable"]
				assert.False(t, hasMax)
				assertYamlIntEquals(t, spec["minAvailable"], 1)
			},
		},
		{
			name: "clusterAgent.pdb.create true (default minAvailable 1)",
			overrides: map[string]string{
				"clusterAgent.pdb.create": "true",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMax := spec["maxUnavailable"]
				assert.False(t, hasMax)
				assertYamlIntEquals(t, spec["minAvailable"], 1)
			},
		},
		{
			name: "clusterAgent.pdb.create true with minAvailable 2",
			overrides: map[string]string{
				"clusterAgent.pdb.create":       "true",
				"clusterAgent.pdb.minAvailable": "2",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMax := spec["maxUnavailable"]
				assert.False(t, hasMax)
				assertYamlIntEquals(t, spec["minAvailable"], 2)
			},
		},
		{
			name: "clusterAgent.pdb.create true with maxUnavailable 3",
			overrides: map[string]string{
				"clusterAgent.pdb.create":         "true",
				"clusterAgent.pdb.maxUnavailable": "3",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMin := spec["minAvailable"]
				assert.False(t, hasMin)
				assertYamlIntEquals(t, spec["maxUnavailable"], 3)
			},
		},
		{
			name: "clusterAgent.pdb.create true fails with both minAvailable and maxUnavailable",
			overrides: map[string]string{
				"clusterAgent.pdb.create":         "true",
				"clusterAgent.pdb.minAvailable":   "1",
				"clusterAgent.pdb.maxUnavailable": "2",
			},
			verify: func(t *testing.T, spec map[string]string) {
				// This test should fail during chart rendering, not reach verification
				t.Fatal("Chart should have failed to render with both minAvailable and maxUnavailable")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-pdb.yaml"},
				Overrides:   tt.overrides,
			})

			// Special handling for the test that should fail
			if strings.Contains(tt.name, "fails with both") {
				assert.NotNil(t, err, "Chart should have failed to render with both minAvailable and maxUnavailable")
				return
			}

			assert.Nil(t, err, "couldn't render template")
			tt.verify(t, extractSpec(manifest))
		})
	}
}

func Test_clusterChecksRunnerPDB(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		verify    func(t *testing.T, spec map[string]string)
	}{
		{
			name: "deprecated clusterChecksRunner.createPodDisruptionBudget true",
			overrides: map[string]string{
				"clusterChecksRunner.createPodDisruptionBudget": "true",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMin := spec["minAvailable"]
				assert.False(t, hasMin)
				assertYamlIntEquals(t, spec["maxUnavailable"], 1)
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true (default maxUnavailable 1)",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create": "true",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMin := spec["minAvailable"]
				assert.False(t, hasMin)
				assertYamlIntEquals(t, spec["maxUnavailable"], 1)
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true with maxUnavailable 2",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create":         "true",
				"clusterChecksRunner.pdb.maxUnavailable": "2",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMin := spec["minAvailable"]
				assert.False(t, hasMin)
				assertYamlIntEquals(t, spec["maxUnavailable"], 2)
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true with minAvailable 1",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create":       "true",
				"clusterChecksRunner.pdb.minAvailable": "1",
			},
			verify: func(t *testing.T, spec map[string]string) {
				_, hasMax := spec["maxUnavailable"]
				assert.False(t, hasMax)
				assertYamlIntEquals(t, spec["minAvailable"], 1)
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true fails with both minAvailable and maxUnavailable",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create":         "true",
				"clusterChecksRunner.pdb.minAvailable":   "1",
				"clusterChecksRunner.pdb.maxUnavailable": "2",
			},
			verify: func(t *testing.T, spec map[string]string) {
				// This test should fail during chart rendering, not reach verification
				t.Fatal("Chart should have failed to render with both minAvailable and maxUnavailable")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/agent-clusterchecks-pdb.yaml"},
				Overrides:   tt.overrides,
			})

			// Special handling for the test that should fail
			if strings.Contains(tt.name, "fails with both") {
				assert.NotNil(t, err, "Chart should have failed to render with both minAvailable and maxUnavailable")
				return
			}

			assert.Nil(t, err, "couldn't render template")
			tt.verify(t, extractSpec(manifest))
		})
	}
}

// extractSpec returns a simplified view of the spec section as key->line string
// so we can do presence/absence checks easily.
func extractSpec(manifest string) map[string]string {
	spec := map[string]string{}
	lines := strings.Split(manifest, "\n")
	inSpec := false
	for _, l := range lines {
		if strings.HasPrefix(l, "spec:") {
			inSpec = true
			continue
		}
		if inSpec {
			// stop at selector (next top-level key in our template)
			if strings.HasPrefix(l, "  selector:") {
				break
			}
			if strings.Contains(l, "minAvailable:") {
				parts := strings.SplitN(l, ":", 2)
				if len(parts) == 2 {
					spec["minAvailable"] = strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(l, "maxUnavailable:") {
				parts := strings.SplitN(l, ":", 2)
				if len(parts) == 2 {
					spec["maxUnavailable"] = strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return spec
}

func assertYamlIntEquals(t *testing.T, got string, expected int) {
	i, err := strconv.Atoi(strings.TrimSpace(got))
	if assert.NoError(t, err) {
		assert.Equal(t, expected, i)
	}
}
