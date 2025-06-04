package abyssal

type HelmRetriever struct {
	ValueSelector string `yaml:"selector"`
	HelmSelector  string `yaml:"entries"`
}