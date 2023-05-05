package e2e

import (
	"context"
	"fmt"
	"github.com/DataDog/helm-charts/test/e2e/pulumi_env"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	if code == 0 {
		fmt.Fprintf(os.Stderr, "Cleaning up stacks")
		errs := pulumi_env.GetStackManager().Cleanup(context.Background())
		for _, err := range errs {
			fmt.Fprint(os.Stderr, err.Error())
		}
	}
	os.Exit(code)
}
