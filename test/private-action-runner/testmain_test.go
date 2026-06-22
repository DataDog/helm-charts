package private_action_runner

import (
	"os"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
)

func TestMain(m *testing.M) {
	common.ParseArgs()
	os.Exit(m.Run())
}
