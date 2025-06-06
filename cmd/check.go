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
		// Your command logic here
		p := providers.NewHelmProvider(Config)

		p.EntriesSelector = Config.Settings.Providers.Helm.EntriesSelector
		p.Directory = "./sample"
		p.Selector = " .cloudimanage.applications "
		p.Evaluator = Config.Settings.Providers.Helm.EvaluatorSelector

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
	},
}

func init() {
	RootCmd.AddCommand(CheckCmd)

	logrus.Debugln("Loaded configuration:", Config)

}
