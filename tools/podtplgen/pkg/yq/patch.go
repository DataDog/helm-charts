package yq

import (
	"fmt"
	"os/exec"
	"strings"
)

type PatchOptions struct {
	Paths []string
}

func Patch(pathFile string, input map[string]interface{}, options PatchOptions) (map[string]interface{}, error) {
	staticArgs := []string{"eval", "-M", "-i"}

	var expressions []string
	for _, yamlPath := range options.Paths {
		expressions = append(expressions, fmt.Sprintf("del(%s)", yamlPath))
	}
	args := append(staticArgs, strings.Join(expressions, "|"), pathFile)
	if err := exec.Command("yq", args...).Run(); err != nil {
		fmt.Printf("error YQ, err: %v", err)
	}

	return nil, nil
}
