package abyssal

/*
  - name: prometheus-stack
    repoUrl: https://prometheus-community.github.io/helm-charts
    chartName: kube-prometheus-stack
    selector: .cloudimanage.applications.{{ .name }} | select(.chartName = \"{{ .chartName }}\") | .targetRevision
    type: helm
    # .entries["kube-prometheus-stack"][0] | .version
    directory: "/"
    schedule:
      - cron: "0 0 * * *"

*/

type AbyssalPackage struct {
	Name      string `yaml:"name"`
	RepoUrl   string `yaml:"repoUrl"`
	ChartName string `yaml:"chartName"`
	Selector  string `yaml:"selector"`
	Type      string `yaml:"type"`
	Directory string `yaml:"directory"`
	Schedule  []struct {
		Cron string `yaml:"cron"`
	} `yaml:"schedule"`
}
