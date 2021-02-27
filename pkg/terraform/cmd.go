package terraform

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

var (
	stdout, stderr bytes.Buffer
)

func terraform(tmpPath string, args ...string) (string, error) {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = tmpPath
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to execute terraform %v", stderr.String())
	}

	return stdout.String(), nil
}

func Init(tmpPath string) error {
	_, err := terraform(tmpPath, "init", "-reconfigure", "-input=false")
	if err != nil {
		return err
	}
	return nil
}

func Apply(tmpPath string) error {
	_, err := terraform(tmpPath, "apply", "-input=false", "-auto-approve", "-lock=false")
	if err != nil {
		return err
	}
	return nil
}

func Output(tmpPath string) (string, error) {
	out, err := terraform(tmpPath, "output", "-json")
	if err != nil {
		return "", err
	}
	return out, nil
}

func Destroy(tmpPath string) error {
	targets := []string{"instance", "bucket"}
	for _, k := range targets {
		_, err := terraform(filepath.Join(tmpPath, k), "destroy", "-input=false", "-auto-approve")
		if err != nil {
			return err
		}
	}
	return nil
}
