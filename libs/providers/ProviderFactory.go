package providers

import (
	"strings"

	"github.com/camalot/abyssal/config"
)

type ProviderFactory struct {

}

func NewProvider(providerElement config.ProviderElement, config *config.AppConfiguration) Provider {
	switch strings.TrimSpace(strings.ToLower(providerElement.Type)) {
	case "argo-aoa":
		return NewArgoAppOfAppsProvider(providerElement, config)
	default:
		return nil
	}
}
