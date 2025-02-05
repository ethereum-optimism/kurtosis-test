package main

import (
	"os"

	"kurtestosis/cli/commands"

	"github.com/sirupsen/logrus"
)

const (
	forceColors   = true
	fullTimestamp = true
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               forceColors,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             fullTimestamp,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})

	err := commands.RootCmd.Execute()
	if err != nil {
		logrus.Debugf("Run exited with error: %v", err)
		
		os.Exit(1)
	}
}