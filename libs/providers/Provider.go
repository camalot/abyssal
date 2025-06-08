package providers

type Provider interface {
	Load() error
	CheckVersionOutOfDate(target ProviderTarget) (bool, string, string, error)
	GetTargets() ([]ProviderTarget, error)
	GetMarkdownTableHeader() string
	GetMarkdownTableRow(target ProviderTarget, outdated bool, currentVersion string, expectedVersion string) string
	GetName() string
}

type ProviderTarget struct {
	Name string `yaml:"name"`
	Source string `yaml:"-"`
	Map  map[string]interface{} `yaml:",inline"`
}
