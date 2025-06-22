package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/camalot/abyssal/libs/notifiers"
	"github.com/camalot/abyssal/libs/providers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ActionCmd = &cobra.Command{
	Use:   "action",
	Short: "Check for outdated packages",
	Long:  `Check for outdated packages in your project. For use from within a github action.`,
	Run: func(cmd *cobra.Command, args []string) {
		Logger.Out = os.Stderr
		start := time.Now()

		fmt.Fprintf(os.Stdout, "# Abyssal - Results\n\n")

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
			fmt.Fprintf(os.Stdout, "## Provider: %s\n\n", p.GetName())

			targets, err := p.GetTargets()
			if err != nil {
				Logger.Fatalf("Error getting targets from provider: %v", err)
			}
			if len(targets) == 0 {
				Logger.Warnf("No targets found for provider %s", provider.Type)
				continue
			}

			fmt.Fprint(os.Stdout, p.GetMarkdownTableHeader())

			for _, target := range targets {
				// convert the targets in to Packages
				result, err := p.CheckVersionOutOfDate(target)
				if err != nil {
					fmt.Fprint(os.Stdout, p.GetMarkdownTableRow(providers.ProviderCheckResult{
						Outdated:        result.Outdated,
						CurrentVersion:  result.CurrentVersion,
						ExpectedVersion: result.ExpectedVersion,
						Target:          target,
						State:           result.State,
						Error:           err.Error(),
					}))
					// write to stderr
					Logger.Out = cmd.ErrOrStderr()
					Logger.Errorf("Error checking package %s: %v", target.Name, err)
					continue
				}
				fmt.Fprint(os.Stdout, p.GetMarkdownTableRow(result))

				for _, notifier := range notifiersList {
					err := notifier.ProcessResult(&result)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error processing result with notifier %s: %v\n", notifier.GetName(), err)
						continue
					}
				}
			}
			fmt.Fprintf(os.Stdout, "%s\n\n", p.GetMarkdownTableFooter())
			fmt.Fprintf(os.Stdout, "%s\n\n", p.GetMarkdownLegend())
		}

		duration := time.Since(start)
		fmt.Fprintf(os.Stdout, "---\n\n")
		fmt.Fprintf(os.Stdout, "Execution Duration: %s\n\n", duration)
	},
}

func init() {
	RootCmd.AddCommand(ActionCmd)

	logrus.Debugln("Loaded configuration:", Config)

}
