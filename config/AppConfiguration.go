package config

import (
	"os"

	"github.com/camalot/abyssal/models/abyssal"
	"gopkg.in/yaml.v3"
)

type AppConfiguration abyssal.Configuration

func Load(filePath string) (AppConfiguration, error) {
	var cfg AppConfiguration

	err := cfg.Load(filePath)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (ac *AppConfiguration) Load(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &ac)
	if err != nil {
		return err
	}

	// merge the default retriever settings with the configuration
	for idx, _:= range ac.Packages {
		pkg := &ac.Packages[idx]
		if pkg.Type == "helm" {
			if pkg.HelmSelector == "" {
				pkg.HelmSelector = ac.Settings.Retrievers.Helm.HelmSelector
			}
			if pkg.ValueSelector == "" {
				pkg.ValueSelector = ac.Settings.Retrievers.Helm.ValueSelector
			}
		}
	}
	return nil
}