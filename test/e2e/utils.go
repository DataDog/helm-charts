package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"os"
)

func SetupConfig() runner.ConfigMap {
	res := runner.ConfigMap{}
	config := os.Getenv("PULUMI_CONFIGS")
	if config != "" {
		var result map[string]map[string]string
		err := json.Unmarshal([]byte(config), &result)
		if err != nil {
			return res
		} else {
			configs := result["config"]
			for key, value := range configs {
				res[key] = auto.ConfigValue{Value: value}
			}
		}
	}
	fmt.Println("RES: ", res)
	return res
}
