package main

import (
	"fmt"
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
	exitCode := extractExitCodeAfterExecution(err)
	os.Exit(exitCode)
}

func extractExitCodeAfterExecution(err error) int {
	if err == nil {
		return 0
	}

	fullErrorMessage := fmt.Sprintf("Error: %v", err)
	commands.RootCmd.PrintErrln(fullErrorMessage)

	return 1
}