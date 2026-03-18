package datadog

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	ddAppsecProxyEnabledEnvVar            = "DD_APPSEC_PROXY_ENABLED"
	ddAppsecProxyAutoDetectEnvVar         = "DD_APPSEC_PROXY_AUTO_DETECT"
	ddAppsecInjectorModeEnvVar            = "DD_CLUSTER_AGENT_APPSEC_INJECTOR_MODE"
	ddAppsecInjectorEnabledEnvVar         = "DD_CLUSTER_AGENT_APPSEC_INJECTOR_ENABLED"
	ddAppsecSidecarImageEnvVar            = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_IMAGE"
	ddAppsecSidecarImageTagEnvVar         = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_IMAGE_TAG"
	ddAppsecSidecarPortEnvVar             = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_PORT"
	ddAppsecSidecarHealthPortEnvVar       = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_HEALTH_PORT"
	ddAppsecSidecarBodyParsingLimitEnvVar = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_BODY_PARSING_SIZE_LIMIT"
	ddAppsecSidecarReqCPUEnvVar           = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_RESOURCES_REQUESTS_CPU"
	ddAppsecSidecarReqMemoryEnvVar        = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_RESOURCES_REQUESTS_MEMORY"
	ddAppsecSidecarLimitCPUEnvVar         = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_RESOURCES_LIMITS_CPU"
	ddAppsecSidecarLimitMemoryEnvVar      = "DD_ADMISSION_CONTROLLER_APPSEC_SIDECAR_RESOURCES_LIMITS_MEMORY"
	ddAppsecProxyProxiesEnvVar            = "DD_APPSEC_PROXY_PROXIES"
	ddAppsecProcessorPortEnvVar           = "DD_APPSEC_PROXY_PROCESSOR_PORT"
	ddAppsecProcessorAddressEnvVar        = "DD_APPSEC_PROXY_PROCESSOR_ADDRESS"
	ddAppsecProcessorServiceNameEnvVar    = "DD_CLUSTER_AGENT_APPSEC_INJECTOR_PROCESSOR_SERVICE_NAME"
	ddAppsecProcessorServiceNsEnvVar      = "DD_CLUSTER_AGENT_APPSEC_INJECTOR_PROCESSOR_SERVICE_NAMESPACE"
)

func renderAppsecInjectorEnvVars(t *testing.T, overrides map[string]string, overridesJSON map[string]string) []corev1.EnvVar {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName:   "datadog",
		ChartPath:     "../../charts/datadog",
		ShowOnly:      []string{"templates/cluster-agent-deployment.yaml"},
		Values:        []string{"../../charts/datadog/values.yaml"},
		Overrides:     overrides,
		OverridesJson: overridesJSON,
	})
	require.NoError(t, err, "failed to render cluster-agent-deployment.yaml")

	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	return deployment.Spec.Template.Spec.Containers[0].Env
}

func Test_AppSecInjector_Disabled_DoesNotRenderAppSecEnvVars(t *testing.T) {
	containerEnv := renderAppsecInjectorEnvVars(t, map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}, nil)

	for _, envVarName := range []string{
		ddAppsecProxyEnabledEnvVar,
		ddAppsecProxyAutoDetectEnvVar,
		ddAppsecInjectorEnabledEnvVar,
		ddAppsecInjectorModeEnvVar,
		ddAppsecProxyProxiesEnvVar,
		ddAppsecProcessorPortEnvVar,
		ddAppsecProcessorAddressEnvVar,
		ddAppsecProcessorServiceNameEnvVar,
		ddAppsecProcessorServiceNsEnvVar,
		ddAppsecSidecarImageEnvVar,
		ddAppsecSidecarImageTagEnvVar,
		ddAppsecSidecarPortEnvVar,
		ddAppsecSidecarHealthPortEnvVar,
		ddAppsecSidecarReqCPUEnvVar,
		ddAppsecSidecarReqMemoryEnvVar,
		ddAppsecSidecarLimitCPUEnvVar,
		ddAppsecSidecarLimitMemoryEnvVar,
		ddAppsecSidecarBodyParsingLimitEnvVar,
	} {
		_, found := findEnvVar(containerEnv, envVarName)
		assert.False(t, found, "did not expect %s when appsec injector is disabled", envVarName)
	}
}

func Test_AppSecInjector_Enabled_RendersDefaultOptions(t *testing.T) {
	containerEnv := renderAppsecInjectorEnvVars(t, map[string]string{
		"datadog.apiKeyExistingSecret":    "datadog-secret",
		"datadog.appKeyExistingSecret":    "datadog-secret",
		"datadog.appsec.injector.enabled": "true",
	}, nil)

	tests := map[string]string{
		ddAppsecProxyEnabledEnvVar:      "true",
		ddAppsecProxyAutoDetectEnvVar:   "true",
		ddAppsecInjectorEnabledEnvVar:   "true",
		ddAppsecInjectorModeEnvVar:      "sidecar",
		ddAppsecSidecarImageEnvVar:      "ghcr.io/datadog/dd-trace-go/service-extensions-callout",
		ddAppsecSidecarImageTagEnvVar:   "v2.6.0",
		ddAppsecSidecarPortEnvVar:       "8080",
		ddAppsecSidecarHealthPortEnvVar: "8081",
		ddAppsecSidecarReqCPUEnvVar:     "10m",
		ddAppsecSidecarReqMemoryEnvVar:  "128Mi",
		ddAppsecProcessorPortEnvVar:     "443",
	}

	for envVarName, expectedValue := range tests {
		envVar, found := findEnvVar(containerEnv, envVarName)
		require.True(t, found, "expected %s to be present", envVarName)
		assert.Equal(t, expectedValue, envVar.Value)
	}

	for _, envVarName := range []string{
		ddAppsecProxyProxiesEnvVar,
		ddAppsecProcessorAddressEnvVar,
		ddAppsecProcessorServiceNameEnvVar,
		ddAppsecProcessorServiceNsEnvVar,
		ddAppsecSidecarLimitCPUEnvVar,
		ddAppsecSidecarLimitMemoryEnvVar,
		ddAppsecSidecarBodyParsingLimitEnvVar,
	} {
		_, found := findEnvVar(containerEnv, envVarName)
		assert.False(t, found, "did not expect %s with default appsec injector values", envVarName)
	}
}

func Test_AppSecInjector_Enabled_RendersCustomOptions(t *testing.T) {
	containerEnv := renderAppsecInjectorEnvVars(t, map[string]string{
		"datadog.apiKeyExistingSecret":                              "datadog-secret",
		"datadog.appKeyExistingSecret":                              "datadog-secret",
		"datadog.appsec.injector.enabled":                           "true",
		"datadog.appsec.injector.autoDetect":                        "false",
		"datadog.appsec.injector.mode":                              "external",
		"datadog.appsec.injector.sidecar.image":                     "ghcr.io/datadog/custom-appsec-sidecar",
		"datadog.appsec.injector.sidecar.imageTag":                  "v2.1.0",
		"datadog.appsec.injector.sidecar.port":                      "18080",
		"datadog.appsec.injector.sidecar.healthPort":                "18081",
		"datadog.appsec.injector.sidecar.bodyParsingSizeLimit":      "10000000",
		"datadog.appsec.injector.sidecar.resources.requests.cpu":    "100m",
		"datadog.appsec.injector.sidecar.resources.requests.memory": "256Mi",
		"datadog.appsec.injector.sidecar.resources.limits.cpu":      "200m",
		"datadog.appsec.injector.sidecar.resources.limits.memory":   "512Mi",
		"datadog.appsec.injector.processor.address":                 "processor.example.svc",
		"datadog.appsec.injector.processor.port":                    "8443",
		"datadog.appsec.injector.processor.service.name":            "appsec-processor",
		"datadog.appsec.injector.processor.service.namespace":       "datadog",
	}, map[string]string{
		"datadog.appsec.injector.proxies": `["envoy-gateway","istio","istio-gateway"]`,
	})

	tests := map[string]string{
		ddAppsecProxyEnabledEnvVar:            "true",
		ddAppsecInjectorEnabledEnvVar:         "true",
		ddAppsecInjectorModeEnvVar:            "external",
		ddAppsecProxyProxiesEnvVar:            "[\"envoy-gateway\",\"istio\",\"istio-gateway\"]",
		ddAppsecProcessorPortEnvVar:           "8443",
		ddAppsecProcessorAddressEnvVar:        "processor.example.svc",
		ddAppsecProcessorServiceNameEnvVar:    "appsec-processor",
		ddAppsecProcessorServiceNsEnvVar:      "datadog",
		ddAppsecSidecarImageEnvVar:            "ghcr.io/datadog/custom-appsec-sidecar",
		ddAppsecSidecarImageTagEnvVar:         "v2.1.0",
		ddAppsecSidecarPortEnvVar:             "18080",
		ddAppsecSidecarHealthPortEnvVar:       "18081",
		ddAppsecSidecarBodyParsingLimitEnvVar: "10000000",
		ddAppsecSidecarReqCPUEnvVar:           "100m",
		ddAppsecSidecarReqMemoryEnvVar:        "256Mi",
		ddAppsecSidecarLimitCPUEnvVar:         "200m",
		ddAppsecSidecarLimitMemoryEnvVar:      "512Mi",
	}

	for envVarName, expectedValue := range tests {
		envVar, found := findEnvVar(containerEnv, envVarName)
		require.True(t, found, "expected %s to be present", envVarName)
		assert.Equal(t, expectedValue, envVar.Value)
	}

	_, found := findEnvVar(containerEnv, ddAppsecProxyAutoDetectEnvVar)
	assert.False(t, found, "did not expect %s when appsec injector autoDetect is disabled", ddAppsecProxyAutoDetectEnvVar)
}

func Test_AppSecInjector_RBAC_IncludesIstioGatewaysRule(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"datadog.apiKeyExistingSecret":    "datadog-secret",
			"datadog.appKeyExistingSecret":    "datadog-secret",
			"datadog.appsec.injector.enabled": "true",
		},
	})
	require.NoError(t, err, "failed to render cluster-agent-rbac.yaml")

	// Find the main cluster-agent ClusterRole from the multi-document manifest.
	var clusterRole rbacv1.ClusterRole
	for _, doc := range strings.Split(manifest, "---") {
		if strings.Contains(doc, "kind: ClusterRole") && strings.Contains(doc, "name: datadog-cluster-agent\n") {
			common.Unmarshal(t, doc, &clusterRole)
			break
		}
	}
	require.NotEmpty(t, clusterRole.Rules, "cluster-agent ClusterRole should have rules")

	var hasEnvoyFiltersRule, hasGatewaysRule bool
	for _, rule := range clusterRole.Rules {
		for _, apiGroup := range rule.APIGroups {
			if apiGroup != "networking.istio.io" {
				continue
			}
			for _, resource := range rule.Resources {
				switch resource {
				case "envoyfilters":
					hasEnvoyFiltersRule = true
				case "gateways":
					hasGatewaysRule = true
					assert.ElementsMatch(t, []string{"get", "list", "watch"}, rule.Verbs,
						"networking.istio.io/gateways rule should have get/list/watch verbs")
				}
			}
		}
	}
	assert.True(t, hasEnvoyFiltersRule, "expected networking.istio.io/envoyfilters rule")
	assert.True(t, hasGatewaysRule, "expected networking.istio.io/gateways rule")
}
