package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfiguration struct {
	Version  string `yaml:"version"`
	Settings struct {
		Notifiers []NotifierElement `yaml:"notifiers"`
		Providers struct {
			ArgoAppOfApps struct {
				EntriesSelector   string `yaml:"entries"`
				EvaluatorSelector string `yaml:"evaluator"`
				BaseSelector      string `yaml:"selector"`
			} `yaml:"argo-aoa"`
		} `yaml:"providers"`
		Authentication map[string]AuthenticationElement `yaml:"authentication"`
	} `yaml:"settings"`
	Providers []ProviderElement `yaml:"providers"`
}

type NotifierElement struct {
	Type    string                 `yaml:"type"`
	Enabled bool                   `yaml:"enabled"`
	Extra   map[string]interface{} `yaml:",inline"`
}

type ProviderElement struct {
	Type  string                 `yaml:"type"`
	Extra map[string]interface{} `yaml:",inline"`
}

type AuthenticationElement struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
	APIKey   string `yaml:"apiKey,omitempty"`
}

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
