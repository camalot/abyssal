package cmd

import (
	// "github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/providers"
	// "github.com/camalot/abyssal/libs/retrievers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for outdated packages",
	Long:  `Check for outdated packages in your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Your command logic here
		p := providers.NewHelmProvider()
		p.HelmSelector = Config.Settings.Providers.Helm.HelmSelector
		p.Directory = "./sample"
		p.Selector = ".. | select(.chartName != null and .repoURL != null) | [{\"chartName\": .chartName, \"repoURL\": .repoURL, \"targetRevision\": .targetRevision}]" // Set the selector for Helm charts
		if err := p.Load(); err != nil {
			Logger.Fatalf("Error loading Helm provider: %v", err)
		}
		for _, target := range p.Targets {
			// convert the targets in to Packages
			outdated, current, expected, err := p.CheckVersionOutOfDate(target)
			if err != nil {
				Logger.Errorf("Error checking package %s: %v", target.ChartName, err)
				continue
			}
			if outdated {
				Logger.Warnf("%s is out of date! Current version: %s - Latest version: %s", target.ChartName, current, expected)
			} else {
				Logger.Infof("%s is up to date.", target.ChartName)
			}

		}
		// r := retrievers.NewHelmRetriever()
		// for _, pkg := range Config.Packages {
		// 	// Check each package for updates
		// 	outdated, current, expected, err := r.CheckVersionOutOfDate(pkg)
		// 	if err != nil {
		// 		Logger.Errorf("Error checking package %s: %v", pkg.Name, err)
		// 		continue
		// 	}
		// 	if outdated {
		// 		Logger.Warnf("%s is out of date! Current version: %s - Latest version: %s", pkg.Name, current, expected)
		// 	} else {
		// 		Logger.Infof("%s is up to date.", pkg.Name)
		// 	}
		//}
	},
}

func init() {
	RootCmd.AddCommand(CheckCmd)

	// Config, err := config.Load(configFile)
	// if err != nil {
	// 	logrus.Fatalln("Error loading config file", err)
	// }

	logrus.Debugln("Loaded configuration:", Config)

}
