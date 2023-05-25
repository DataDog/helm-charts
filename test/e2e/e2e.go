package e2e

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"
	"os"
)

func teardownSuite() {
	fmt.Fprintf(os.Stderr, "Cleaning up stacks. ")
	errs := infra.GetStackManager().Cleanup(context.Background())
	for _, err := range errs {
		fmt.Fprint(os.Stderr, err.Error())
	}
}
