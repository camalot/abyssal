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

	return nil
}