// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/require"
)

// TestMapperNegativeCases verifies that the mapper correctly rejects invalid configurations.
func TestMapperNegativeCases(t *testing.T) {
	for _, tc := range negativeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := runMapperExpectError(t, tc.ValuesFile)
			require.Error(t, err, "Expected mapper to return an error for %s, but it succeeded", tc.Name)

			if tc.ExpectedErrMsg != "" {
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.ExpectedErrMsg),
					"Error message should contain expected substring for %s", tc.Name)
			}

			t.Logf("Mapper correctly returned error: %v", err)
		})
	}
}

// TestInvalidYAMLChartInstall verifies that invalid YAML causes Helm chart installation to fail.
func TestInvalidYAMLChartInstall(t *testing.T) {
	invalidValuesFile := negativeValuesDir + "/invalid-yaml-values.yaml"
	_, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: releaseDatadog,
		ChartPath:   datadogChartPath,
		Values:      []string{invalidValuesFile},
	})

	require.Error(t, err, "Helm chart should fail to render with invalid YAML values")
	t.Logf("Helm chart correctly rejected invalid YAML: %v", err)
}

