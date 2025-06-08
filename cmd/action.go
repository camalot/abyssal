package cmd

import (
	"fmt"
	"time"

	"github.com/camalot/abyssal/libs/providers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ActionCmd = &cobra.Command{
	Use:   "action",
	Short: "Check for outdated packages",
	Long:  `Check for outdated packages in your project. For use from within a github action.`,
	Run: func(cmd *cobra.Command, args []string) {
		Logger.Out = cmd.ErrOrStderr()
		start := time.Now()

		fmt.Printf("# Abyssal - Results\n\n")

		for _, provider := range Config.Providers {
			p := providers.NewProvider(provider, Config)
			if err := p.Load(); err != nil {
				Logger.Fatalf("Error loading Argo provider: %v", err)
			}
			targets, err := p.GetTargets()
			if err != nil {
				Logger.Fatalf("Error getting targets from provider: %v", err)
			}
			if len(targets) == 0 {
				Logger.Warnf("No targets found for provider %s", provider.Type)
				continue
			}

			fmt.Printf("| Package | Current Version | Expected Version | Status |\n")
			fmt.Printf("|---------|-----------------|------------------|--------|\n")

			for _, target := range targets {
				// convert the targets in to Packages
				outdated, current, expected, err := p.CheckVersionOutOfDate(target)
				if err != nil {
					fmt.Printf("| %s | Error checking package: %v |\n", target.Name, err)
					// write to stderr
					Logger.Errorf("Error checking package %s: %v", target.Name, err)
					continue
				}
				if outdated {
					fmt.Printf("| %s | %s | %s | ❌ |\n", target.Name, current, expected)
				} else {
					fmt.Printf("| %s | %s | %s | ✅ |\n", target.Name, current, expected)
				}
			}
		}

		duration := time.Since(start)
		fmt.Printf("\n---\n\n")
		fmt.Printf("Execution Duration: %s\n", duration)
	},
}

func init() {
	RootCmd.AddCommand(ActionCmd)

	logrus.Debugln("Loaded configuration:", Config)

}
