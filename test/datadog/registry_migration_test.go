package datadog

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

// TestRegistryMigration tests the registry helper under all combinations of site,
// migration mode, and relevant overrides (APM, GKE Autopilot/GDC, explicit registry).
func TestRegistryMigration(t *testing.T) {
	// Site × mode matrix.
	// In auto mode, AP1 migrates because datadog.apm.enabled defaults to false.
	sites := []struct {
		name         string
		site         string // empty = default (datadoghq.com / US1)
		wantAuto     string
		wantAll      string
		wantDisabled string
	}{
		{
			// apm.enabled defaults to false, so auto mode migrates US1.
			name:         "US1 (default)",
			wantAuto:     "registry.datadoghq.com",
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
			wantAuto:     "registry.datadoghq.com",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "gcr.io/datadoghq",
		},
		{
			name:         "EU1",
			site:         "datadoghq.eu",
			wantAuto:     "registry.datadoghq.com",
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
			// apm.enabled defaults to false, so auto mode migrates AP1.
			name:         "AP1",
			site:         "ap1.datadoghq.com",
			wantAuto:     "registry.datadoghq.com",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "asia.gcr.io/datadoghq",
		},
		{
			name:         "AP2",
			site:         "ap2.datadoghq.com",
			wantAuto:     "registry.datadoghq.com",
			wantAll:      "registry.datadoghq.com",
			wantDisabled: "gcr.io/datadoghq",
		},
	}

	modes := []struct {
		name  string
		value string
		want  func(s struct {
			name, site, wantAuto, wantAll, wantDisabled string
		}) string
	}{
		{"auto", "auto", func(s struct{ name, site, wantAuto, wantAll, wantDisabled string }) string {
			return s.wantAuto
		}},
		{"all", "all", func(s struct{ name, site, wantAuto, wantAll, wantDisabled string }) string {
			return s.wantAll
		}},
		{"disabled", "", func(s struct{ name, site, wantAuto, wantAll, wantDisabled string }) string {
			return s.wantDisabled
		}},
	}

	for _, site := range sites {
		t.Run(site.name, func(t *testing.T) {
			for _, mode := range modes {
				t.Run("mode="+mode.name, func(t *testing.T) {
					overrides := map[string]string{
						"datadog.apiKeyExistingSecret": "datadog-secret",
						"datadog.appKeyExistingSecret": "datadog-secret",
						"registryMigrationMode":        mode.value,
					}
					if site.site != "" {
						overrides["datadog.site"] = site.site
					}
					assert.Equal(t, mode.want(site), renderAndExtractRegistry(t, overrides))
				})
			}
		})
	}

	// Invalid registryMigrationMode values must be rejected with an error.
	t.Run("invalid mode: fails fast", func(t *testing.T) {
		_, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"registryMigrationMode":        "Auto",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "registryMigrationMode")
	})

	// AP1 auto migration applies regardless of APM configuration.
	t.Run("AP1/auto/apm-enabled: migrates", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.site":                 "ap1.datadoghq.com",
			"datadog.apm.enabled":          "true",
			"registryMigrationMode":        "auto",
		})
		assert.Equal(t, "registry.datadoghq.com", registry)
	})

	// US1 auto migration is gated on APM being disabled (both legacy and modern fields).
	t.Run("US1/auto/apm-enabled: does not migrate", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.apm.enabled":          "true",
			"registryMigrationMode":        "auto",
		})
		assert.Equal(t, "gcr.io/datadoghq", registry)
	})

	t.Run("US1/auto/apm-portEnabled: does not migrate", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.apm.portEnabled":      "true",
			"registryMigrationMode":        "auto",
		})
		assert.Equal(t, "gcr.io/datadoghq", registry)
	})

	// Explicit registry always takes precedence over migration.
	t.Run("explicit registry overrides migration", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.site":                 "ap1.datadoghq.com",
			"registryMigrationMode":        "auto",
			"registry":                     "my-custom-registry.example.com",
		})
		assert.Equal(t, "my-custom-registry.example.com", registry)
	})

	// GKE GDC on US3 should fall through to gcr.io, not datadoghq.azurecr.io.
	t.Run("US3/GKE GDC: uses gcr.io not azurecr", func(t *testing.T) {
		registry := renderAndExtractRegistry(t, map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"datadog.site":                 "us3.datadoghq.com",
			"registryMigrationMode":        "auto",
			"providers.gke.gdc":            "true",
		})
		assert.Equal(t, "gcr.io/datadoghq", registry)
	})

	// GKE Autopilot and GKE GDC always bypass migration, even with mode=all.
	for _, provider := range []struct {
		name string
		key  string
	}{
		{"GKE Autopilot", "providers.gke.autopilot"},
		{"GKE GDC", "providers.gke.gdc"},
	} {
		t.Run(provider.name+"/ap1/auto: bypasses migration", func(t *testing.T) {
			registry := renderAndExtractRegistry(t, map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"datadog.site":                 "ap1.datadoghq.com",
				"registryMigrationMode":        "auto",
				provider.key:                   "true",
			})
			assert.Equal(t, "asia.gcr.io/datadoghq", registry)
		})

		t.Run(provider.name+"/default/all: bypasses migration", func(t *testing.T) {
			registry := renderAndExtractRegistry(t, map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"registryMigrationMode":        "all",
				provider.key:                   "true",
			})
			assert.Equal(t, "gcr.io/datadoghq", registry)
		})
	}
}

// TestAdmissionControllerContainerRegistry verifies that DD_ADMISSION_CONTROLLER_CONTAINER_REGISTRY
// is excluded from migration and always uses site-specific registries regardless of registryMigrationMode.
func TestAdmissionControllerContainerRegistry(t *testing.T) {
	tests := []struct {
		name          string
		site          string
		mode          string
		wantRegistry  string
	}{
		// Migration must not affect the admission controller registry.
		{name: "US1/auto", site: "", mode: "auto", wantRegistry: "gcr.io/datadoghq"},
		{name: "US1/all", site: "", mode: "all", wantRegistry: "gcr.io/datadoghq"},
		{name: "EU1/auto", site: "datadoghq.eu", mode: "auto", wantRegistry: "eu.gcr.io/datadoghq"},
		{name: "EU1/all", site: "datadoghq.eu", mode: "all", wantRegistry: "eu.gcr.io/datadoghq"},
		{name: "AP1/auto", site: "ap1.datadoghq.com", mode: "auto", wantRegistry: "asia.gcr.io/datadoghq"},
		{name: "AP1/all", site: "ap1.datadoghq.com", mode: "all", wantRegistry: "asia.gcr.io/datadoghq"},
		{name: "US5/auto", site: "us5.datadoghq.com", mode: "auto", wantRegistry: "gcr.io/datadoghq"},
		{name: "US5/all", site: "us5.datadoghq.com", mode: "all", wantRegistry: "gcr.io/datadoghq"},
		// Explicit containerRegistry override takes precedence.
		{name: "explicit override", site: "", mode: "auto", wantRegistry: "my-custom-registry.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"clusterAgent.enabled":         "true",
				"registryMigrationMode":        tt.mode,
			}
			if tt.site != "" {
				overrides["datadog.site"] = tt.site
			}
			if tt.name == "explicit override" {
				overrides["clusterAgent.admissionController.containerRegistry"] = "my-custom-registry.example.com"
			}
			assert.Equal(t, tt.wantRegistry, renderAndExtractAdmissionControllerRegistry(t, overrides))
		})
	}
}

func renderAndExtractAdmissionControllerRegistry(t *testing.T, overrides map[string]string) string {
	t.Helper()
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	require.NoError(t, err, "couldn't render template")

	var deploy appsv1.Deployment
	common.Unmarshal(t, manifest, &deploy)

	for _, container := range deploy.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == "DD_ADMISSION_CONTROLLER_CONTAINER_REGISTRY" {
				return env.Value
			}
		}
	}
	t.Fatal("DD_ADMISSION_CONTROLLER_CONTAINER_REGISTRY not found in cluster-agent deployment")
	return ""
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
