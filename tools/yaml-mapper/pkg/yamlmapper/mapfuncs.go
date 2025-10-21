package yamlmapper

import (
	"log"
	"reflect"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

var CustomMapFuncs = map[string]customMapFunc{
	"mapApiSecretKey":        mapApiSecretKey,
	"mapAppSecretKey":        mapAppSecretKey,
	"mapTokenSecretKey":      mapTokenSecretKey,
	"mapSeccompProfile":      mapSeccompProfile,
	"mapSystemProbeAppArmor": mapSystemProbeAppArmor,
	"mapLocalServiceName":    mapLocalServiceName,
	"mapAppendEnvVar":        mapAppendEnvVar,
	"mapMergeEnvs":           mapMergeEnvs,
}

type customMapFunc func(values map[string]interface{}, newPath string, pathVal interface{}, args []interface{})

func mapApiSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	//	if existing apikey secret, need to add key-name
	setInterim(interim, newPath, pathVal)
	interim["spec.global.credentials.apiSecret.keyName"] = "api-key"
}

func mapAppSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	setInterim(interim, newPath, pathVal)
	interim["spec.global.credentials.appSecret.keyName"] = "app-key"
}

func mapTokenSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	setInterim(interim, newPath, pathVal)
	interim["spec.global.clusterAgentTokenSecret.keyName"] = "token"
}

func mapSeccompProfile(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	seccompValue, err := pathVal.(string)
	if !err {
		return
	}

	if strings.HasPrefix(seccompValue, "localhost/") {
		profileName := strings.TrimPrefix(seccompValue, "localhost/")
		setInterim(interim, newPath+".type", "Localhost")
		setInterim(interim, newPath+".localhostProfile", profileName)

	} else if seccompValue == "runtime/default" {
		setInterim(interim, newPath+".type", "RuntimeDefault")

	} else if seccompValue == "unconfined" {
		setInterim(interim, newPath+".type", "Unconfined")

	}
}

func mapSystemProbeAppArmor(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	appArmorValue, err := pathVal.(string)
	if !err || appArmorValue == "" {
		// must be set to non-empty string
		return
	}

	systemProbeFeatures := []string{
		"spec.features.cws.enabled",            // datadog.securityAgent.runtime.enabled
		"spec.features.npm.enabled",            // datadog.networkMonitoring.enabled
		"spec.features.tcpQueueLength.enabled", // datadog.systemProbe.enableTCPQueueLength
		"spec.features.oomKill.enabled",        // datadog.systemProbe.enableOOMKill
		"spec.features.usm.enabled",            // datadog.serviceMonitoring.enabled
	}

	hasSystemProbeFeature := false
	for _, feature := range systemProbeFeatures {
		if val, exists := interim[feature]; exists {
			if enabled, ok := val.(bool); ok && enabled {
				hasSystemProbeFeature = true
				break
			}
		}
	}

	if !hasSystemProbeFeature {
		gpuEnabled, gpuExists := interim["spec.features.gpu.enabled"]
		gpuPrivileged, privExists := interim["spec.features.gpu.privilegedMode"]
		if gpuExists && privExists {
			if gpuEnabledBool, ok := gpuEnabled.(bool); ok && gpuEnabledBool {
				if gpuPrivilegedBool, ok := gpuPrivileged.(bool); ok && gpuPrivilegedBool {
					hasSystemProbeFeature = true
				}
			}
		}
	}

	if hasSystemProbeFeature {
		// must be set to non-empty string
		setInterim(interim, newPath, appArmorValue)
	}
}

func mapLocalServiceName(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	nameOverride, ok := pathVal.(string)
	if !ok || nameOverride == "" {
		return
	}
	setInterim(interim, newPath, nameOverride)
}

// mapAppendEnvVar appends environment variables to a specified path in the interim configuration.
// It takes a list of environment variable definitions in the format []map[string]interface{}{{"name": "VAR_NAME"}}
// and creates new environment variable entries with the provided pathVal as the value.
// The new variables are added to the interim map at the specified newPath.
// Example:
//   - mapFuncArgs: []interface{}{map[string]interface{}{"name": "DD_LOG_LEVEL"}}
//   - pathVal: "debug"
//   - Result: Appends {"name": "DD_LOG_LEVEL", "value": "debug"} to newPath in interim
func mapAppendEnvVar(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	if len(args) != 1 {
		return
	}

	envMap, ok := args[0].(map[string]interface{})
	if !ok {
		log.Printf("expected map[string]interface{} for env var map definition, got %T", args[0])
		return
	}
	newEnvName := envMap["name"].(string)
	newEnvVar := map[string]interface{}{
		"name":  newEnvName,
		"value": pathVal,
	}

	// Handle valueFrom
	pathValType := reflect.TypeOf(pathVal).Kind().String()

	if pathValType == "map" || pathValType == "string" {
		var valFrom interface{}
		var ok bool
		var strOk bool
		var pathValStr string

		_, valOk := pathVal.(chartutil.Values)
		_, mapOk := pathVal.(map[string]interface{})

		if pathValType == "string" {
			pathValStr, _ = pathVal.(string)
			strOk = strings.Contains(pathValStr, "valueFrom")
		}

		if valOk {
			valFrom, ok = pathVal.(chartutil.Values)["valueFrom"]
		} else if mapOk {
			valFrom, ok = pathVal.(map[string]interface{})["valueFrom"]
		} else if strOk {
			var data map[string]interface{}
			err := yaml.Unmarshal([]byte(pathValStr), &data)
			if err == nil {
				valFrom = data["valueFrom"]
				ok = true
			}
		}

		if ok {
			newEnvVar = map[string]interface{}{
				"name":      envMap["name"],
				"valueFrom": valFrom,
			}
		}
	}

	// Create the interim[newPath] if it doesn't exist yet
	if _, exists := interim[newPath]; !exists {
		setInterim(interim, newPath, []interface{}{newEnvVar})
		return
	}

	existingEnvs, ok := interim[newPath].([]interface{})
	if !ok {
		log.Printf("Error: expected []interface{} at path %s, got %T", newPath, interim[newPath])
		return
	}

	envExists := hasDuplicateEnv(existingEnvs, newEnvName)

	if !envExists {
		setInterim(interim, newPath, append(existingEnvs, newEnvVar))
	}
}

// mapMergeEnvs merges lists of environment variables at the specified path.
// It takes a slice of environment variable maps and merges them with any existing
// environment variables at the target path.
// Example:
//   - pathVal: []map[string]interface{}{{"name": "VAR1", "value": "val1"}}
//   - Result: Merges the new env vars with any existing ones at newPath
func mapMergeEnvs(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	newEnvs, ok := pathVal.([]interface{})
	if !ok {
		log.Printf("Warning: expected []interface{} for pathVal, got %T", pathVal)
		return
	}

	//Initialize mergedEnvs with existing environments or an empty slice
	var mergedEnvs []interface{}
	if existingEnvs, exists := interim[newPath]; exists {
		if existingEnvsSlice, ok := existingEnvs.([]interface{}); ok {
			// Make a copy of existing environments
			mergedEnvs = make([]interface{}, len(existingEnvsSlice))
			copy(mergedEnvs, existingEnvsSlice)
		}
	}

	// Add new envs that don't already exist
	for _, newEnv := range newEnvs {
		newEnvMap, ok := newEnv.(map[string]interface{})
		if !ok {
			log.Printf("Warning: expected map[string]interface{} in newEnvs, got %T", newEnv)
			continue
		}

		newName, ok := newEnvMap["name"].(string)
		if !ok || newName == "" {
			log.Printf("Warning: missing or invalid 'name' field in environment variable: %v", newEnvMap)
			continue
		}

		if !hasDuplicateEnv(mergedEnvs, newName) {
			mergedEnvs = append(mergedEnvs, newEnv)
		} else {
			// Override existing env with new env
			for i, existingEnv := range mergedEnvs {
				if existingMap, ok := existingEnv.(map[string]interface{}); ok {
					if existingName, ok := existingMap["name"].(string); ok && existingName == newName {
						mergedEnvs[i] = newEnv
					}
				}
			}
		}
	}

	if len(mergedEnvs) > 0 {
		setInterim(interim, newPath, mergedEnvs)
	}
}

func hasDuplicateEnv(existingEnvs []interface{}, newEnvName interface{}) bool {
	envExists := false
	for _, existingEnv := range existingEnvs {
		if existingMap, ok := existingEnv.(map[string]interface{}); ok {
			if existingName, ok := existingMap["name"].(string); ok && existingName == newEnvName {
				envExists = true
				break
			}
		}
	}
	return envExists
}
