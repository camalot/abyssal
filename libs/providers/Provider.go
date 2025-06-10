package providers

import (
	"strings"
)

type ProviderCheckResult struct {
	Outdated        bool           `json:"outdated" yaml:"outdated"`
	Target          ProviderTarget `json:"target" yaml:"target"`
	CurrentVersion  string         `json:"current_version" yaml:"current_version"`
	ExpectedVersion string         `json:"expected_version" yaml:"expected_version"`
	Error           string         `json:"error,omitempty" yaml:"error,omitempty"`
	State           ProviderState  `json:"state" yaml:"state"`
}

type ProviderState string

const (
	ProviderCheckStateSuccess ProviderState = "success"
	ProviderCheckStateSkipped ProviderState = "skipped"
	ProviderCheckStateError   ProviderState = "error"
)

type ProviderStateEmoji string

const (
	ProviderCheckStateEmojiSuccess ProviderStateEmoji = "✅"
	ProviderCheckStateEmojiSkipped ProviderStateEmoji = "⏭️"
	ProviderCheckStateEmojiFailure ProviderStateEmoji = "❌"
	ProviderCheckStateEmojiError   ProviderStateEmoji = "⚠️"
	ProviderCheckStateEmojiUnknown ProviderStateEmoji = "❓"
)

type Provider interface {
	Load() error
	CheckVersionOutOfDate(target ProviderTarget) (ProviderCheckResult, error)
	GetTargets() ([]ProviderTarget, error)
	// GenerateMarkdown(result ProviderCheckResult) string
	GetMarkdownTableHeader() string
	GetMarkdownTableRow(result ProviderCheckResult) string
	GetName() string
	GetMarkdownLegend() string
	GetMarkdownTableFooter() string
}

type ProviderTarget struct {
	Name   string                 `yaml:"name"`
	Source string                 `yaml:"-"`
	Map    map[string]interface{} `yaml:",inline"`
}

func cleanValueForVersion(value string) string {
	// Remove any leading or trailing whitespace
	value = strings.TrimSpace(value)
	// Remove any leading 'v' character
	value = strings.TrimPrefix(value, "v")
	// remove any quotes
	value = strings.Trim(value, "\"'")

	// Remove any trailing characters that are not digits, dots, or hyphens
	// for i := len(value) - 1; i >= 0; i-- {
	// 	if !(value[i] >= '0' && value[i] <= '9') && value[i] != '.' && value[i] != '-' {
	// 		value = value[:i]
	// 	} else {
	// 		break
	// 	}
	// }

	return value
}
