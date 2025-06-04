package abyssal

type Configuration struct {
	Version  string           `yaml:"version"`
	Packages []AbyssalPackage `yaml:"packages"`
	Settings struct {
		Notifiers struct {
			Github  map[string]interface{} `yaml:"github"`
			Slack   map[string]interface{} `yaml:"slack"`
			Discord map[string]interface{} `yaml:"discord"`
			Email   map[string]interface{} `yaml:"email"`
			Jira    map[string]interface{} `yaml:"jira"`
		} `yaml:"notifiers"`
		Retrievers struct {
			Helm HelmRetriever `yaml:"helm"`
		} `yaml:"retrievers"`
	} `yaml:"settings"`
}
