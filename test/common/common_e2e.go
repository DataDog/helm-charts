package common

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

var (
	defaultAgentVersion  = "latest"
	agentVersion         = os.Getenv("E2E_AGENT_VERSION")
	clusterAgentVersion  = agentVersion
	defaultImageRegistry = "gcr.io/datadoghq"
	defaultPulumiConfigs = runner.ConfigMap{
		"ddinfra:kubernetesVersion": auto.ConfigValue{Value: "1.32"},
	}

	defaultCIPulumiConfigs = runner.ConfigMap{
		"ddinfra:env":                           auto.ConfigValue{Value: "gcp/agent-qa"},
		"ddinfra:gcp/defaultPrivateKeyPassword": auto.ConfigValue{Value: os.Getenv("E2E_GCP_PRIVATE_KEY_PASSWORD"), Secret: true},
	}
)

func parseE2EConfigParams() []string {
	// "key1=val1 key2=val2"
	configParams := os.Getenv("E2E_CONFIG_PARAMS")
	if len(configParams) < 1 {
		return []string{}
	}
	// ["key1=val1", "key2=val2"]
	configKVs := strings.Split(configParams, " ")
	return configKVs
}

// SetupConfig test
func SetupConfig() (runner.ConfigMap, error) {
	res := runner.ConfigMap{}
	configs := parseE2EConfigParams()

	if agentVersion == "" {
		agentVersion = defaultAgentVersion
		clusterAgentVersion = defaultAgentVersion
	}

	// DCA release candidates are tagged as "rc" in the registry.
	if clusterAgentVersion == "7-rc" {
		clusterAgentVersion = "rc"
	}

	agentImageConfigs := runner.ConfigMap{
		"ddagent:fullImagePath":             auto.ConfigValue{Value: fmt.Sprintf("%s/agent:%s", defaultImageRegistry, agentVersion)},
		"ddagent:clusterAgentFullImagePath": auto.ConfigValue{Value: fmt.Sprintf("%s/cluster-agent:%s", defaultImageRegistry, clusterAgentVersion)},
	}
	defaultPulumiConfigs.Merge(agentImageConfigs)

	if os.Getenv("E2E_PROFILE") == "ci" {
		res.Merge(defaultPulumiConfigs)
		res.Merge(defaultCIPulumiConfigs)
	} else {
		// use "local" E2E profile for local testing
		// fast-fail if missing required env vars
		_, e2eApiKeyBool := os.LookupEnv("E2E_API_KEY")
		_, e2eAppKeyBool := os.LookupEnv("E2E_APP_KEY")
		_, e2eAwsKeypairNameBool := os.LookupEnv("E2E_KEY_PAIR_NAME")
		if !e2eApiKeyBool || !e2eAppKeyBool || !e2eAwsKeypairNameBool {
			return nil, fmt.Errorf("missing required environment variables. Must set `E2E_API_KEY`, `E2E_APP_KEY`, and `E2E_KEY_PAIR_NAME` for the local E2E profile")
		} else {
			res.Merge(defaultPulumiConfigs)
		}
	}

	if len(configs) > 0 {
		for _, config := range configs {
			kv := strings.Split(config, "=")
			if _, exists := res[kv[0]]; !exists {
				isSecret := strings.Contains(strings.ToLower(kv[0]), "password")
				res[kv[0]] = auto.ConfigValue{Value: kv[1], Secret: isSecret}
			} else {
				log.Printf("Config param %s used more than once.", kv[0])
			}
		}
	}
	return res, nil
}
