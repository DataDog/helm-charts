package helm

import (
	"context"
	"os/exec"

	yaml "gopkg.in/yaml.v2"
)

type RenderOptions struct {
	MandatoryOptions []string
	Options          []string
	ChartName        string
	OutputFilePath   string
}

func Render(ctx context.Context, options RenderOptions) (map[string]interface{}, error) {
	args := []string{"template", "-s", "templates/daemonset.yaml"}
	for _, opt := range append(options.MandatoryOptions, options.Options...) {
		if opt == "" {
			continue
		}
		args = append(args, "--set", opt)
	}
	args = append(args, options.ChartName)
	cmd := exec.CommandContext(ctx, "helm", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(stdout)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}
