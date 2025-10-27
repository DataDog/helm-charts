package main

const (
	confdValue = `
redisdb.yaml: |-
  init_config:
  instances:
    - host: "name"
	  port: "6379"
`

	checksdValue = `
	service.py: |-

`
)

var allowOptions = []string{
	"datadog.kubeStateMetricsCore.enabled=true",
	"datadog.logs.enabled=true",
	"datadog.logs.containerCollectAll=true",
	"datadog.apm.portEnabled=true",
	"clusterAgent.metricsProvider.enabled=true",
	"datadog.processAgent.enabled=true",
	"datadog.containerRuntimeSupport.enabled=false",
	"agents.useConfigMap=true",
	"datadog.confd=" + confdValue,
	"datadog.checksd=" + checksdValue,
}

var removePaths = []string{
	".metadata",
	".spec.template.metadata",
	".spec.template.spec.affinity",
	".spec.template.spec.containers[].env",
	".spec.template.spec.containers[].image",
	".spec.template.spec.containers[].imagePullPolicy",
	".spec.template.spec.containers[].resources",
	".spec.template.spec.initContainers[].env",
	".spec.template.spec.initContainers[].image",
	".spec.template.spec.initContainers[].imagePullPolicy",
	".spec.template.spec.initContainers[].resources",
	".spec.template.spec.tolerations",
	".spec.template.spec.tolerations",
	".spec.template.spec.nodeSelector",
	".spec.updateStrategy",
	".spec.selector",
}
