package datadog_operator

import (
	"flag"
	"os"
	"testing"
)

var UpdateBaselines bool

func TestMain(m *testing.M) {
	flag.BoolVar(&UpdateBaselines, "updateBaselines", false, "When set to true overwrites existing baselines with the rendered ones")
	flag.Parse()
	os.Exit(m.Run())
}
