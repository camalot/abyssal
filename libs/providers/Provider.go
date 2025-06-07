package providers

type Provider interface {
	Load() error
	CheckVersionOutOfDate(target ProviderTarget) (bool, string, string, error)
	GetTargets() ([]ProviderTarget, error)
}

type ProviderTarget struct {
	Name string `yaml:"name"`
	Map  map[string]interface{} `yaml:",inline"`
}
