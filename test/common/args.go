package common

import "flag"

var UpdateBaselines bool
var PreserveStacks bool
var DestroyStacks bool

func ParseArgs() {
	flag.BoolVar(&UpdateBaselines, "updateBaselines", false, "When set to true overwrites existing baselines with the rendered ones")
	flag.BoolVar(&PreserveStacks, "preserveStacks", false, "When set to true, preserves newly-created or existing Pulumi end-to-end (E2E) stacks after completing tests. When set to false, destroys Pulumi E2E stacks upon test completion.")
	flag.BoolVar(&DestroyStacks, "destroyStacks", false, "When set to true, destroys existing Pulumi end-to-end (E2E) stacks and performs cleanup. When set to false, `preserveStacks` value takes precedence.")
	flag.Parse()
}
