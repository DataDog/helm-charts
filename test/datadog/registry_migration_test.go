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
	sites := []struct {
		name         string
		site         string // empty = default (datadoghq.com)
		wantAuto     string // expected registry in "auto" mode
		wantAll      string // expected registry in "all" mode
		wantDisabled string // expected registry when migration is disabled ("")
	}{
		{
			name:         "US1 (default)",
			site:         "",
			wantAuto:     "gcr.io/datadoghq",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "gcr.io/datadoghq",
		},
		{
			name:         "US3",
			site:         "us3.datadoghq.com",
			wantAuto:     "datadoghq.azurecr.io",
			wantAll:      "datadoghq.azurecr.io",
			wantDisabled: "datadoghq.azurecr.io",
		},
		{
			name:         "US5",
			site:         "us5.datadoghq.com",
			wantAuto:     "gcr.io/datadoghq",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "gcr.io/datadoghq",
		},
		{
			name:         "EU1",
			site:         "datadoghq.eu",
			wantAuto:     "eu.gcr.io/datadoghq",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "eu.gcr.io/datadoghq",
		},
		{
			name:         "US1-FED",
			site:         "ddog-gov.com",
			wantAuto:     "public.ecr.aws/datadog",
			wantAll:      "public.ecr.aws/datadog",
			wantDisabled: "public.ecr.aws/datadog",
		},
		{
			name:         "AP1",
			site:         "ap1.datadoghq.com",
			wantAuto:     "registry.datadoghq.com",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "asia.gcr.io/datadoghq",
		},
		{
			name:         "AP2",
			site:         "ap2.datadoghq.com",
			wantAuto:     "gcr.io/datadoghq",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "gcr.io/datadoghq",
		},
	}

	modes := []struct {
		name  string
		value string
	}{
		{"auto", "auto"},
		{"all", "all"},
		{"disabled", ""},
	}

	for _, site := range sites {
		t.Run(site.name, func(t *testing.T) {
			for _, mode := range modes {
				var expected string
				switch mode.name {
				case "auto":
					expected = site.wantAuto
				case "all":
					expected = site.wantAll
				case "disabled":
					expected = site.wantDisabled
				}

				t.Run("mode="+mode.name, func(t *testing.T) {
					overrides := map[string]string{
						"datadog.apiKeyExistingSecret": "datadog-secret",
						"datadog.appKeyExistingSecret": "datadog-secret",
						"registryMigrationMode":        mode.value,
					}
					if site.site != "" {
						overrides["datadog.site"] = site.site
					}

					registry := renderAndExtractRegistry(t, overrides)
					assert.Equal(t, expected, registry)
				})
			}
		})
	}
}

func TestRegistryMigrationOverrides(t *testing.T) {
	t.Run("explicit registry takes precedence", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.site":                 "ap1.datadoghq.com",
			"registryMigrationMode":        "auto",
			"registry":                     "my-custom-registry.example.com",
		})
		assert.Equal(t, "my-custom-registry.example.com", registry)
	})

	t.Run("GKE Autopilot bypasses migration for ap1", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.site":                 "ap1.datadoghq.com",
			"registryMigrationMode":        "auto",
			"providers.gke.autopilot":      "true",
		})
		assert.Equal(t, "asia.gcr.io/datadoghq", registry)
	})

	t.Run("GKE Autopilot bypasses migration for default site", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"registryMigrationMode":        "all",
			"providers.gke.autopilot":      "true",
		})
		assert.Equal(t, "gcr.io/datadoghq", registry)
	})
}

func renderAndExtractRegistry(t *testing.T, overrides map[string]string) string {
	t.Helper()
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
	parts := strings.SplitN(agentImage, "/", 3)
	if len(parts) == 3 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}
