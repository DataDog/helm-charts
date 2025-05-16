package common

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

var defaultPulumiConfigs = runner.ConfigMap{
	"ddinfra:aws/defaultKeyPairName": auto.ConfigValue{Value: os.Getenv("E2E_KEY_PAIR_NAME")},
}
var defaultCIPulumiConfigs = runner.ConfigMap{
	"aws:skipCredentialsValidation":     auto.ConfigValue{Value: "true"},
	"aws:skipMetadataApiCheck":          auto.ConfigValue{Value: "false"},
	"ddinfra:aws/defaultPrivateKeyPath": auto.ConfigValue{Value: os.Getenv("E2E_AWS_PRIVATE_KEY_PATH")},
}

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

func SetupConfig() (runner.ConfigMap, error) {
	res := runner.ConfigMap{}
	configs := parseE2EConfigParams()
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
			_, exists := res[kv[0]]
			if !exists {
				res[kv[0]] = auto.ConfigValue{Value: kv[1]}
			} else {
				log.Printf("Config param %s used more than once. Value: %s", kv[0], kv[1])
			}
		}
	}
	log.Printf("Setting up Pulumi E2E stack with configs: %v", res)
	return res, nil
}
