package cmd

import (
	"fmt"
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
		Logger.Out = cmd.OutOrStderr()
		start := time.Now()

		fmt.Printf("# Abyssal - Results\n\n")

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
			fmt.Printf("## Provider: %s\n\n", p.GetName())

			targets, err := p.GetTargets()
			if err != nil {
				Logger.Fatalf("Error getting targets from provider: %v", err)
			}
			if len(targets) == 0 {
				Logger.Warnf("No targets found for provider %s", provider.Type)
				continue
			}

			fmt.Print(p.GetMarkdownTableHeader())

			for _, target := range targets {
				// convert the targets in to Packages
				result, err := p.CheckVersionOutOfDate(target)
				if err != nil {
					fmt.Print(p.GetMarkdownTableRow(providers.ProviderCheckResult{
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
				fmt.Print(p.GetMarkdownTableRow(result))

				for _, notifier := range notifiersList {
					notifyPayload, err := notifier.CreatePayload(notifier.GetNotifierConfig(), &result)
					if err != nil {
						Logger.Errorf("Error creating payload for notifier %s: %v", notifier.GetName(), err)
						continue
					}
					if result.Outdated && notifier.NeedsNotification(notifyPayload) {
						if err := notifier.Notify(notifyPayload); err != nil {
							Logger.Errorf("Error sending notification with %s: %v", notifier.GetName(), err)
						} else {
							Logger.Infof("Notification sent with %s for package %s", notifier.GetName(), target.Name)
						}
					}
				}
			}
			fmt.Printf("%s\n\n", p.GetMarkdownTableFooter())
			fmt.Printf("%s\n\n", p.GetMarkdownLegend())
		}

		duration := time.Since(start)
		fmt.Printf("---\n\n")
		fmt.Printf("Execution Duration: %s\n\n", duration)
	},
}

func init() {
	RootCmd.AddCommand(ActionCmd)

	logrus.Debugln("Loaded configuration:", Config)

}
