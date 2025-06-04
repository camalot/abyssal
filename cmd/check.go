package cmd

import (
	// "github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/retrievers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for outdated packages",
	Long:  `Check for outdated packages in your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Your command logic here

		r := retrievers.NewHelmRetriever()
		for _, pkg := range Config.Packages {
			logrus.Infof("Checking package: %s", pkg.Name)
			// Check each package for updates
			outdated, result, err := r.OutOfDateVersion(pkg)
			if err != nil {
				logrus.Errorf("Error checking package %s: %v", pkg.Name, err)
				continue
			}
			if outdated {
				logrus.Warnf("Package %s is out of date! Latest version: %s", pkg.Name, result)
			} else {
				logrus.Infof("Package %s is up to date.", pkg.Name)
			}
		}
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