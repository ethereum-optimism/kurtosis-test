package main

import (
	"os"

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

	exitCode := 0

	logrus.Info("kurtestosis CLI is still work in progress")

	os.Exit(exitCode)
}