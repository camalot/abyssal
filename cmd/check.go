package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/camalot/abyssal/libs/notifiers"
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

		Logger.Out = os.Stderr
		start := time.Now()

		// prepare notifiers
		notifiersList := make([]notifiers.Notifier, 0)
		if len(Config.Settings.Notifiers) >= 0 {
			for _, notifier := range Config.Settings.Notifiers {
				n := notifiers.NewNotifier(notifier, Config)
				if n == nil {
					Logger.Warnf("Notifier type %s is not supported or not implemented", notifier.Type)
					continue
				}
				if !n.IsEnabled() {
					Logger.Debugf("Notifier %s is disabled, skipping", n.GetName())
					continue
				}
				notifiersList = append(notifiersList, n)
			}
		}

		for _, provider := range Config.Providers {

			p := providers.NewProvider(provider, Config)
			if err := p.Load(); err != nil {
				Logger.Fatalf("Error loading Argo provider: %v", err)
			}
			fmt.Fprintf(os.Stdout, "Provider: %s\n", p.GetName())

			targets, err := p.GetTargets()
			if err != nil {
				Logger.Fatalf("Error getting targets from provider: %v", err)
			}
			if len(targets) == 0 {
				Logger.Warnf("No targets found for provider %s", provider.Type)
				continue
			}

			for _, target := range targets {
				fmt.Fprintf(os.Stdout, "\nTarget: %s\n", target.Name)
				fmt.Fprintf(os.Stdout, "Source: %s\n", target.Source)
				// convert the targets in to Packages
				result, err := p.CheckVersionOutOfDate(target)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error checking package %s: %v\n", target.Name, err)
					continue
				}

				for _, notifier := range notifiersList {
					err := notifier.ProcessResult(&result)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error processing result with notifier %s: %v\n", notifier.GetName(), err)
						continue
					}
				}
			}
		}

		duration := time.Since(start)
		fmt.Fprintf(os.Stdout, "Execution Duration: %s\n\n", duration)
	},
}

func init() {
	RootCmd.AddCommand(CheckCmd)

	CheckCmd.PersistentFlags().StringVarP(&outputFormat, "output-format", "o", "text", "Output format (text|json|yaml|markdown)")
	CheckCmd.PersistentFlags().Lookup("output-format").DefValue = "text"
	viper.BindPFlag("output-format", CheckCmd.PersistentFlags().Lookup("output-format"))

	logrus.Debugln("Loaded configuration:", Config)

}
