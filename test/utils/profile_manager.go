package utils

import (
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner/parameters"
	"os"
	"os/user"
	"strings"
)

const (
	EnvPrefix = "E2E_"
	envSep    = ","
)

type Profile runner.Profile

// Shared implementations for common profiles methods
type baseProfile struct {
	projectName  string
	environments []string
	store        parameters.Store
	secretStore  parameters.Store
}

type localProfile struct {
	baseProfile
}

type ciProfile struct {
	baseProfile
	ciUniqueID string
}

func (p baseProfile) EnvironmentNames() string {
	return strings.Join(p.environments, envSep)
}

func (p baseProfile) ProjectName() string {
	return p.projectName
}

func (p ciProfile) RootWorkspacePath() string {
	return workspaceFolder
}

func (p localProfile) RootWorkspacePath() string {
	return workspaceFolder
}

func (p baseProfile) ParamStore() parameters.Store {
	return p.store
}

func (p baseProfile) SecretStore() parameters.Store {
	return p.secretStore
}

func (p ciProfile) NamePrefix() string {
	return p.ciUniqueID
}

func (p localProfile) NamePrefix() string {
	// Stack names may only contain alphanumeric characters, hyphens, underscores, or periods.
	// As NamePrefix is used as stack name, we sanitize the user name.
	var username string
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}

	if username == "" || username == "root" {
		username = "nouser"
	}

	parts := strings.Split(username, ".")
	if numParts := len(parts); numParts > 1 {
		var usernameBuilder strings.Builder
		for _, part := range parts[0 : numParts-1] {
			usernameBuilder.WriteByte(part[0])
		}
		usernameBuilder.WriteString(parts[numParts-1])
		username = usernameBuilder.String()
	}

	username = strings.ToLower(username)
	username = strings.ReplaceAll(username, " ", "-")

	return username
}

func (p baseProfile) AllowDevMode() bool {
	return true
}

func (p ciProfile) AllowDevMode() bool {
	return false
}

func newProfile(projectName string, environments []string, secretStore *parameters.Store) baseProfile {
	p := baseProfile{
		projectName:  projectName,
		environments: environments,
		store:        parameters.NewEnvStore(EnvPrefix),
	}

	if secretStore == nil {
		p.secretStore = p.store
	} else {
		p.secretStore = *secretStore
	}

	return p
}

func NewCIProfile() (Profile, error) {
	// Create workspace directory
	if err := os.MkdirAll(workspaceFolder, 0o700); err != nil {
		return nil, fmt.Errorf("unable to create temporary folder at: %s, err: %w", workspaceFolder, err)
	}

	// Secret store
	secretStore := parameters.NewAWSStore("ci.helm-charts.")

	// Set Pulumi password
	passVal, err := secretStore.Get(parameters.PulumiPassword)
	if err != nil {
		return nil, fmt.Errorf("unable to get pulumi state password, err: %w", err)
	}
	os.Setenv("PULUMI_CONFIG_PASSPHRASE", passVal)

	// Building name prefix
	pipelineID := os.Getenv("CI_PIPELINE_ID")
	projectID := os.Getenv("CI_PROJECT_ID")
	if pipelineID == "" || projectID == "" {
		return nil, fmt.Errorf("unable to compute name prefix, missing variables pipeline id: %s, project id: %s", pipelineID, projectID)
	}

	return ciProfile{
		baseProfile: newProfile("helm-charts-e2eci", []string{"aws/sandbox"}, &secretStore),
		ciUniqueID:  pipelineID + "-" + projectID,
	}, nil
}

func GetProfile() Profile {
	return runner.GetProfile()
}
