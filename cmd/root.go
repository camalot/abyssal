package cmd

import (
	"fmt"

	"github.com/camalot/abyssal/config"
	"github.com/camalot/abyssal/libs/envs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var (
	configFile string
	logLevel   string
	noColor    bool

	Config *config.AppConfiguration
	Logger *logrus.Logger
)

var RootCmd = &cobra.Command{
	Use:   "abyssal",
	Short: "Abyssal CLI - a commandline tool for tracking out of date packages.",
	Long:  `Abyssal CLI is a command-line interface for tracking out of date packages in your projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Your command logic here
		cmd.Help() // Display help if no subcommand is provided
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load environment variables from .env file
		envs.LoadDotEnv("./.env")

		Config = &config.AppConfiguration{}
		err := Config.Load(configFile)
		if err != nil {
			logrus.Fatalln("Error loading config file", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Cleanup or finalization logic after all commands have run
	},
}

func init() {
	var err error
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./.abyssal.yaml", "config file (default is .abyssal.yaml)")



	RootCmd.PersistentFlags().BoolVarP(&noColor, "no-color", "!", false, "Disable color output")
	viper.BindPFlag("no-color", RootCmd.PersistentFlags().Lookup("no-color"))

	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "debug", "set the log level (debug, info, warn, error, fatal, panic)")
	viper.BindPFlag("log-level", RootCmd.PersistentFlags().Lookup("log-level"))
	logLevel, err := logrus.ParseLevel(viper.GetString("log-level"))

	Logger = &logrus.Logger{
		Out:   logrus.StandardLogger().Out,
		Level: logLevel,
		Formatter: &easy.Formatter{
			LogFormat: "[%lvl%] %msg%\n",
		},
	}

	if err != nil {
		Logger.Fatalln("Error parsing log level", err)
	}
}
