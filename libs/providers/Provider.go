package providers


type ProviderCheckResult struct {
	Outdated bool   `json:"outdated" yaml:"outdated"`
	Target ProviderTarget `json:"target" yaml:"target"`
	CurrentVersion string `json:"current_version" yaml:"current_version"`
	ExpectedVersion string `json:"expected_version" yaml:"expected_version"`
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
	State ProviderState `json:"state" yaml:"state"`
}

type ProviderState string

const (
	ProviderCheckStateSuccess   ProviderState = "success"
	ProviderCheckStateFailure   ProviderState = "failure"
	ProviderCheckStateSkipped   ProviderState = "skipped"
	ProviderCheckStateWarning   ProviderState = "warning"
	ProviderCheckStateError     ProviderState = "error"
)

type Provider interface {
	Load() error
	CheckVersionOutOfDate(target ProviderTarget) (ProviderCheckResult, error)
	GetTargets() ([]ProviderTarget, error)
	// GenerateMarkdown(result ProviderCheckResult) string
	GetMarkdownTableHeader() string
	GetMarkdownTableRow(result ProviderCheckResult) string
	GetName() string
}

type ProviderTarget struct {
	Name string `yaml:"name"`
	Source string `yaml:"-"`
	Map  map[string]interface{} `yaml:",inline"`
}
