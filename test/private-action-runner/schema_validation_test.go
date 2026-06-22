package private_action_runner

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/assert"
)

func TestPreviousValuesSchema(t *testing.T) {

	command := common.HelmCommand{
		ReleaseName: "0-dot-x-values-file",
		ChartPath:   "../../charts/private-action-runner",
		Values:      []string{"./data/old-values-file.yaml"},
	}
	_, err := common.RenderChart(t, command)
	assert.NotNil(t, err, "expect schema validation error")
	stderr := err.(*shell.ErrWithCmdOutput).Output.Stderr()
	assert.Contains(t, stderr, "Error: values don't meet the specifications of the schema(s) in the following chart(s):")
}
