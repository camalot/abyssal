package cmd

import (
	"github.com/camalot/abyssal/libs/providers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	outputFormat string
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
	},
}

func init() {
	RootCmd.AddCommand(CheckCmd)

	CheckCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "text", "Output format (text|json|yaml|markdown)")
	CheckCmd.PersistentFlags().Lookup("output-format").DefValue = "text"
	viper.BindPFlag("output-format", CheckCmd.PersistentFlags().Lookup("output-format"))

	logrus.Debugln("Loaded configuration:", Config)

}
