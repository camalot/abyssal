package config

import (
	"os"

	"github.com/camalot/abyssal/models/abyssal"
	"gopkg.in/yaml.v3"
)

type PackageConfiguration abyssal.AbyssalConfiguration

func Load(filePath string) (PackageConfiguration, error) {
	var cfg PackageConfiguration

	err := cfg.Load(filePath)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (pc *PackageConfiguration) Load(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &pc)
	if err != nil {
		return err
	}

	return nil
}