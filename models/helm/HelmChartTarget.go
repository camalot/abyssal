package helm

type HelmChartTarget struct {
	ChartName      string `yaml:"chartName"`
	RepoURL        string `yaml:"repoURL"`
	TargetRevision string `yaml:"targetRevision"`
}
