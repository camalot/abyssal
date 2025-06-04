package main

import (
	"github.com/camalot/abyssal/cmd"
	"github.com/sirupsen/logrus"
)

func main() {

  if err := cmd.RootCmd.Execute(); err != nil {
    if logrus.GetLevel() < logrus.WarnLevel {
      logrus.Errorln("debug error by increasing log level (e.g. debug)")
    }
    logrus.WithField("extended", err.Error()).
      Fatalln("an error occurred executing the command")
  }
}
