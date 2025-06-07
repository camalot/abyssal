package cmd

import (
	"github.com/camalot/abyssal/libs/providers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for outdated packages",
	Long:  `Check for outdated packages in your project.`,
	Run: func(cmd *cobra.Command, args []string) {

		for _, provider := range Config.Providers {
			p := providers.NewProvider(provider, Config)
			if err := p.Load(); err != nil {
				Logger.Fatalf("Error loading Argo provider: %v", err)
			}
			targets, err := p.GetTargets()
			if err != nil {
				Logger.Fatalf("Error getting targets from provider: %v", err)
			}
			for _, target := range targets {
				// convert the targets in to Packages
				outdated, current, expected, err := p.CheckVersionOutOfDate(target)
				if err != nil {
					Logger.Errorf("Error checking package %s: %v", target.Name, err)
					continue
				}
				if outdated {
					Logger.Warnf("%s is out of date! Current version: %s - Latest version: %s", target.Name, current, expected)
				} else {
					Logger.Infof("%s is up to date.", target.Name)
				}

			}
		}

		// TODO: loop the providers and load them accordingly
		// p := providers.NewArgoAppOfAppsProvider(Config)

		// p.EntriesSelector = Config.Settings.Providers.ArgoAppOfApps.EntriesSelector
		// p.Directory = "./sample"
		// p.Selector = " .cloudimanage.applications "
		// p.Evaluator = Config.Settings.Providers.ArgoAppOfApps.EvaluatorSelector


	},
}

func init() {
	RootCmd.AddCommand(CheckCmd)

	logrus.Debugln("Loaded configuration:", Config)

}
