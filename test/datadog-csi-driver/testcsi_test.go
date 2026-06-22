package datadog_csi_driver

import (
	"os"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
)

func TestMain(m *testing.M) {
	common.ParseArgs()
	os.Exit(m.Run())
}
