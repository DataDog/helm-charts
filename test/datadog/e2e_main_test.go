package datadog

import (
	"flag"
	"os"
	"testing"
)

var PreserveStacks bool

func TestMain(m *testing.M) {
	flag.BoolVar(&PreserveStacks, "preserveStacks", false, "When set to true, preserves newly-created or existing Pulumi end-to-end (E2E) stacks after completing tests. When set to false, destroys Pulumi E2E stacks upon test completion.")
	flag.Parse()
	os.Exit(m.Run())
}
