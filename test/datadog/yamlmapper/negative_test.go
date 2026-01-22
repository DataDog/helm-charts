// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"os"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Negative Test Cases
// =============================================================================

// TestMapperNegativeCases verifies that the mapper correctly rejects invalid configurations.
// These tests ensure that unsupported or invalid Helm keys are not silently ignored.
func TestMapperNegativeCases(t *testing.T) {
	// Ensure negative test directory exists
	if _, err := os.Stat(negativeValuesDir); os.IsNotExist(err) {
		t.Skipf("Negative test values directory does not exist: %s", negativeValuesDir)
	}

	for _, tc := range negativeTestCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			// Check if values file exists
			if _, err := os.Stat(tc.ValuesFile); os.IsNotExist(err) {
				t.Skipf("Negative test values file does not exist: %s", tc.ValuesFile)
			}

			t.Logf("Testing: %s", tc.Description)

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
// This is a sanity check to ensure the Datadog chart properly validates its input.
func TestInvalidYAMLChartInstall(t *testing.T) {
	// Skip if negative values directory doesn't exist
	if _, err := os.Stat(negativeValuesDir); os.IsNotExist(err) {
		t.Skipf("Negative test values directory does not exist: %s", negativeValuesDir)
	}

	invalidValuesFile := negativeValuesDir + "/invalid-yaml-values.yaml"
	if _, err := os.Stat(invalidValuesFile); os.IsNotExist(err) {
		t.Skipf("Invalid YAML test file does not exist: %s", invalidValuesFile)
	}

	// Attempt to render the chart with invalid values - this should fail
	_, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: releaseDatadog,
		ChartPath:   datadogChartPath,
		Values:      []string{invalidValuesFile},
	})

	require.Error(t, err, "Helm chart should fail to render with invalid YAML values")
	t.Logf("Helm chart correctly rejected invalid YAML: %v", err)
}

