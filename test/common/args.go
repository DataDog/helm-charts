package common

import "flag"

var UpdateBaselines bool

func ParseArgs() {
	flag.BoolVar(&UpdateBaselines, "updateBaselines", false, "When set to true overwrites existing baselines with the rendered ones")
	flag.Parse()
}
