package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_clusterAgentPDB(t *testing.T) {
	tests := []struct {
		name       string
		overrides  map[string]string
		expectFail bool
		verify     func(t *testing.T, pdb policyv1.PodDisruptionBudget)
	}{
		{
			name: "deprecated clusterAgent.createPodDisruptionBudget true",
			overrides: map[string]string{
				"clusterAgent.createPodDisruptionBudget": "true",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MaxUnavailable)
				assert.NotNil(t, pdb.Spec.MinAvailable)
				assert.Equal(t, intstr.Int, pdb.Spec.MinAvailable.Type)
				assert.Equal(t, 1, pdb.Spec.MinAvailable.IntValue())
			},
		},
		{
			name: "clusterAgent.pdb.create true (default minAvailable 1)",
			overrides: map[string]string{
				"clusterAgent.pdb.create": "true",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MaxUnavailable)
				assert.NotNil(t, pdb.Spec.MinAvailable)
				assert.Equal(t, 1, pdb.Spec.MinAvailable.IntValue())
			},
		},
		{
			name: "clusterAgent.pdb.create true with minAvailable 20%",
			overrides: map[string]string{
				"clusterAgent.pdb.create":       "true",
				"clusterAgent.pdb.minAvailable": "20%",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MaxUnavailable)
				assert.Equal(t, "20%", pdb.Spec.MinAvailable.StrVal)
			},
		},
		{
			name: "clusterAgent.pdb.create true with maxUnavailable 3",
			overrides: map[string]string{
				"clusterAgent.pdb.create":         "true",
				"clusterAgent.pdb.maxUnavailable": "3",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MinAvailable)
				assert.Equal(t, 3, pdb.Spec.MaxUnavailable.IntValue())
			},
		},
		{
			name: "clusterAgent.pdb.create true fails with both minAvailable and maxUnavailable",
			overrides: map[string]string{
				"clusterAgent.pdb.create":         "true",
				"clusterAgent.pdb.minAvailable":   "1",
				"clusterAgent.pdb.maxUnavailable": "2",
			},
			expectFail: true,
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

			if tt.expectFail {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			var pdb policyv1.PodDisruptionBudget
			common.Unmarshal(t, manifest, &pdb)
			tt.verify(t, pdb)
		})
	}
}

func Test_clusterChecksRunnerPDB(t *testing.T) {
	tests := []struct {
		name       string
		overrides  map[string]string
		expectFail bool
		verify     func(t *testing.T, pdb policyv1.PodDisruptionBudget)
	}{
		{
			name: "deprecated clusterChecksRunner.createPodDisruptionBudget true",
			overrides: map[string]string{
				"clusterChecksRunner.createPodDisruptionBudget": "true",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MinAvailable)
				assert.Equal(t, 1, pdb.Spec.MaxUnavailable.IntValue())
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true (default maxUnavailable 1)",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create": "true",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MinAvailable)
				assert.Equal(t, 1, pdb.Spec.MaxUnavailable.IntValue())
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true with maxUnavailable 2",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create":         "true",
				"clusterChecksRunner.pdb.maxUnavailable": "2",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MinAvailable)
				assert.Equal(t, 2, pdb.Spec.MaxUnavailable.IntValue())
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true with minAvailable 1",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create":       "true",
				"clusterChecksRunner.pdb.minAvailable": "1",
			},
			verify: func(t *testing.T, pdb policyv1.PodDisruptionBudget) {
				assert.Nil(t, pdb.Spec.MaxUnavailable)
				assert.Equal(t, 1, pdb.Spec.MinAvailable.IntValue())
			},
		},
		{
			name: "clusterChecksRunner.pdb.create true fails with both minAvailable and maxUnavailable",
			overrides: map[string]string{
				"clusterChecksRunner.pdb.create":         "true",
				"clusterChecksRunner.pdb.minAvailable":   "1",
				"clusterChecksRunner.pdb.maxUnavailable": "2",
			},
			expectFail: true,
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

			if tt.expectFail {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			var pdb policyv1.PodDisruptionBudget
			common.Unmarshal(t, manifest, &pdb)
			tt.verify(t, pdb)
		})
	}
}
