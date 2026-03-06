package datadog

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

func TestRegistryMigrationMode(t *testing.T) {
	tests := []struct {
		name             string
		overrides        map[string]string
		expectedRegistry string
	}{
		{
			name: "auto mode with ap1 site uses registry.datadoghq.com",
			overrides: map[string]string{
				"datadog.site":          "ap1.datadoghq.com",
				"registryMigrationMode": "auto",
			},
			expectedRegistry: "registry.datadoghq.com",
		},
		{
			name: "auto mode with default site (us1) keeps gcr.io",
			overrides: map[string]string{
				"registryMigrationMode": "auto",
			},
			expectedRegistry: "gcr.io/datadoghq",
		},
		{
			name: "auto mode with eu site keeps eu.gcr.io",
			overrides: map[string]string{
				"datadog.site":          "datadoghq.eu",
				"registryMigrationMode": "auto",
			},
			expectedRegistry: "eu.gcr.io/datadoghq",
		},
		{
			name: "all mode with default site (us1) uses registry.datadoghq.com",
			overrides: map[string]string{
				"registryMigrationMode": "all",
			},
			expectedRegistry: "registry.datadoghq.com",
		},
		{
			name: "all mode with eu site uses registry.datadoghq.com",
			overrides: map[string]string{
				"datadog.site":          "datadoghq.eu",
				"registryMigrationMode": "all",
			},
			expectedRegistry: "registry.datadoghq.com",
		},
		{
			name: "all mode with us3 site keeps datadoghq.azurecr.io",
			overrides: map[string]string{
				"datadog.site":          "us3.datadoghq.com",
				"registryMigrationMode": "all",
			},
			expectedRegistry: "datadoghq.azurecr.io",
		},
		{
			name: "all mode with ap1 site uses registry.datadoghq.com",
			overrides: map[string]string{
				"datadog.site":          "ap1.datadoghq.com",
				"registryMigrationMode": "all",
			},
			expectedRegistry: "registry.datadoghq.com",
		},
		{
			name: "disabled mode with ap1 site keeps asia.gcr.io",
			overrides: map[string]string{
				"datadog.site":          "ap1.datadoghq.com",
				"registryMigrationMode": "",
			},
			expectedRegistry: "asia.gcr.io/datadoghq",
		},
		{
			name: "disabled mode with default site keeps gcr.io",
			overrides: map[string]string{
				"registryMigrationMode": "",
			},
			expectedRegistry: "gcr.io/datadoghq",
		},
		{
			name: "gov site always uses ecr regardless of migration mode",
			overrides: map[string]string{
				"datadog.site":          "ddog-gov.com",
				"registryMigrationMode": "all",
			},
			expectedRegistry: "public.ecr.aws/datadog",
		},
		{
			name: "explicit registry overrides migration mode",
			overrides: map[string]string{
				"datadog.site":          "ap1.datadoghq.com",
				"registryMigrationMode": "auto",
				"registry":              "my-custom-registry.example.com",
			},
			expectedRegistry: "my-custom-registry.example.com",
		},
		{
			name: "gke autopilot with ap1 keeps asia.gcr.io despite auto mode",
			overrides: map[string]string{
				"datadog.site":            "ap1.datadoghq.com",
				"registryMigrationMode":   "auto",
				"providers.gke.autopilot": "true",
			},
			expectedRegistry: "asia.gcr.io/datadoghq",
		},
		{
			name: "gke autopilot with all mode keeps asia.gcr.io for ap1",
			overrides: map[string]string{
				"datadog.site":            "ap1.datadoghq.com",
				"registryMigrationMode":   "all",
				"providers.gke.autopilot": "true",
			},
			expectedRegistry: "asia.gcr.io/datadoghq",
		},
		{
			name: "gke autopilot with all mode keeps gcr.io for default site",
			overrides: map[string]string{
				"registryMigrationMode":   "all",
				"providers.gke.autopilot": "true",
			},
			expectedRegistry: "gcr.io/datadoghq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
			}
			for k, v := range tt.overrides {
				overrides[k] = v
			}

			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.NoError(t, err, "couldn't render template")

			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)

			agentImage := ds.Spec.Template.Spec.Containers[0].Image
			registry := strings.Split(agentImage, "/")[0]
			if strings.Count(agentImage, "/") > 1 {
				// Handle registries like eu.gcr.io/datadoghq or public.ecr.aws/datadog
				parts := strings.SplitN(agentImage, "/", 3)
				registry = parts[0] + "/" + parts[1]
			}
			assert.Equal(t, tt.expectedRegistry, registry, "unexpected registry in image: %s", agentImage)
		})
	}
}
