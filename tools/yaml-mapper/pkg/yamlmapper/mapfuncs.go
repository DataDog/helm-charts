package yamlmapper

import "strings"

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
	interim[newPath] = pathVal
	interim["spec.global.credentials.apiSecret.keyName"] = "api-key"
}

func mapAppSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	interim[newPath] = pathVal
	interim["spec.global.credentials.appSecret.keyName"] = "app-key"
}

func mapTokenSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	interim[newPath] = pathVal
	interim["spec.global.clusterAgentTokenSecret.keyName"] = "token"
}

func mapSeccompProfile(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	seccompValue, err := pathVal.(string)
	if !err {
		return
	}

	if strings.HasPrefix(seccompValue, "localhost/") {
		profileName := strings.TrimPrefix(seccompValue, "localhost/")
		interim[newPath+".type"] = "Localhost"
		interim[newPath+".localhostProfile"] = profileName

	} else if seccompValue == "runtime/default" {
		interim[newPath+".type"] = "RuntimeDefault"

	} else if seccompValue == "unconfined" {
		interim[newPath+".type"] = "Unconfined"

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
		interim[newPath] = appArmorValue
	}
}

func mapLocalServiceName(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	nameOverride, ok := pathVal.(string)
	if !ok || nameOverride == "" {
		return
	}
	interim[newPath] = nameOverride
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
		//log.Printf("expected map[string]interface{} for env var map definition, got %T", args[0])
		return
	}

	newEnvVar := map[string]interface{}{
		"name":  envMap["name"],
		"value": pathVal,
	}

	// Create the interim[newPath] if it doesn't exist yet
	if _, exists := interim[newPath]; !exists {
		interim[newPath] = []interface{}{newEnvVar}
		return
	}

	existing, ok := interim[newPath].([]interface{})
	if !ok {
		//log.Printf("Error: expected []interface{} at path %s, got %T", newPath, interim[newPath])
		return
	}

	interim[newPath] = append(existing, newEnvVar)
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
		//log.Printf("Warning: expected []interface{} for pathVal, got %T", pathVal)
		return
	}

	// If the interim[newPath] doesn't exist yet, just set the new environment variables
	existingEnvs, exists := interim[newPath]
	if !exists {
		interim[newPath] = newEnvs
		return
	}

	existingEnvsSlice, ok := existingEnvs.([]interface{})
	if !ok {
		//log.Printf("Warning: expected []interface{} at path %s, got %T", newPath, existingEnvs)
		return
	}

	// Merge the slices, avoiding duplicates
	mergedEnvs := make([]interface{}, len(existingEnvsSlice))
	copy(mergedEnvs, existingEnvsSlice)

	// Add new envs that don't already exist
	for _, newEnv := range newEnvs {
		newEnvMap, ok := newEnv.(map[string]interface{})
		if !ok {
			//log.Printf("Warning: expected map[string]interface{} in newEnvs, got %T", newEnv)
			continue
		}

		exists := false
		newName, hasName := newEnvMap["name"].(string)
		if !hasName {
			continue
		}

		for _, existingEnv := range mergedEnvs {
			if existingMap, ok := existingEnv.(map[string]interface{}); ok {
				if existingName, ok := existingMap["name"].(string); ok && existingName == newName {
					exists = true
					break
				}
			}
		}

		if !exists {
			mergedEnvs = append(mergedEnvs, newEnv)
		}
	}

	interim[newPath] = mergedEnvs
}
