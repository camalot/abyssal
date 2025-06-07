package providers

import "github.com/camalot/abyssal/config"

type ProviderFactory struct {

}

func NewProvider(providerElement config.ProviderElement, config *config.AppConfiguration) Provider {
	switch providerElement.Type {
	case "argo-aoa":
		return NewArgoAppOfAppsProvider(providerElement, config)

	// case "helm":
	// 	return NewHelmProvider()
	// case "kustomize":
	// 	return NewKustomizeProvider()
	default:
		return nil
	}
}
