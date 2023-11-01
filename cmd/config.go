package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/a8m/envsubst"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"boundless-cli/pkg/config"
)

const KubeConfigFile = "kubeconfig"

func initBlueprint(ctx *cli.Context) (config.Blueprint, error) {
	f := ctx.String("config")
	if f == "" {
		f = "blueprint.yaml"
	}

	file, err := configReader(f)
	if err != nil {
		return config.Blueprint{}, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return config.Blueprint{}, err
	}

	subst, err := envsubst.Bytes(content)
	if err != nil {
		return config.Blueprint{}, err
	}

	log.Debugf("Loaded configuration:\n%s", subst)
	cfg, err := config.ParseBoundlessCluster(subst)
	if err != nil {
		return config.Blueprint{}, err
	}

	return cfg, nil
}

func configReader(f string) (io.ReadCloser, error) {
	if f == "-" {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return nil, fmt.Errorf("can't stat stdin: %s", err.Error())
		}
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return os.Stdin, nil
		}
		return nil, fmt.Errorf("can't read stdin")
	}

	variants := []string{f}
	// add .yml to default value lookup
	if f == "k0sctl.yaml" {
		variants = append(variants, "k0sctl.yml")
	}

	for _, fn := range variants {
		if _, err := os.Stat(fn); err != nil {
			continue
		}

		fp, err := filepath.Abs(fn)
		if err != nil {
			return nil, err
		}
		file, err := os.Open(fp)
		if err != nil {
			return nil, err
		}

		return file, nil
	}

	return nil, fmt.Errorf("failed to locate configuration")
}

func createKubeConfig(kubeconfig string) error {
	c := exec.Command("k0sctl", "kubeconfig", "--config", kubeconfig)
	c.Stderr = os.Stderr

	buf := new(bytes.Buffer)
	c.Stdout = buf

	err := c.Run()
	if err != nil {
		return fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	return os.WriteFile(KubeConfigFile, buf.Bytes(), 0600)
}

func getK0sctlConfigPath(c *cli.Context) (string, error) {
	blueprint, err := initBlueprint(c)
	if err != nil {
		return "", err
	}

	k0sctlConfig := config.ConvertToK0s(blueprint)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(k0sctlConfig)
	if err != nil {
		return "", err
	}

	k0sctlConfigFile, err := writeToTempFile(data)
	if err != nil {
		return "", err
	}

	return k0sctlConfigFile, nil
}

func writeToTempFile(data []byte) (string, error) {
	// create tmp file for k0sctl config file
	tmpfile, err := os.CreateTemp("", "k0sctl.yaml")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = tmpfile.Write(data)
	if err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}
