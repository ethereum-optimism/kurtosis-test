package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	logLevelStrFlag = "log-level"
)

var logLevelStr string
var defaultLogLevelStr = logrus.InfoLevel.String()

// RootCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var RootCmd = &cobra.Command{
	Use:   KurtestosisCmdStr,
	Short: "Kurtestosis, Kurtosis test runner CLI",
	// Cobra will print usage whenever _any_ error occurs, including ones we throw in Kurtosis
	// This doesn't make sense in 99% of the cases, so just turn them off entirely
	SilenceUsage: true,
	// Cobra prints the errors itself, however, with this flag disabled it give Kurtosis control
	// and allows us to post process the error in the main.go file.
	SilenceErrors:     true,
	PersistentPreRunE: setupCLI,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: run,
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&logLevelStr,
		logLevelStrFlag,
		defaultLogLevelStr,
		"Sets the level that the CLI will log at ("+strings.Join(getAllLogLevelStrings(), "|")+")",
	)
}

func run(cmd *cobra.Command, args []string) error {
	logrus.Warn("kurtestosis CLI is still work in progress")

	globPattern := args[0]

	// First we expand the glob into file paths
	// 
	// These are the test suites that we will run
    testSuitePaths, err := filepath.Glob(globPattern)
    if err != nil {
        logrus.Errorf("Error expanding glob pattern: %v", err)

		return err
    }

	// Exit if there are no test suites to run
    if len(testSuitePaths) == 0 {
        logrus.Warn("No test suites found matching the glob pattern")
        
		return nil
    }

	// Talk to the user
	logrus.Debugf("Found %d matching test suites:\n%s", len(testSuitePaths), strings.Join(testSuitePaths, "\n"))

	return nil
}

// Concatenates all logrus log level strings into a string array
func getAllLogLevelStrings() []string {
	result := []string{}
	for _, level := range logrus.AllLevels {
		levelStr := level.String()
		result = append(result, levelStr)
	}
	return result
}

// Setup function to run before any command execution
func setupCLI(cmd *cobra.Command, args []string) error {
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return fmt.Errorf("error parsing the %s CLI argument: %w", logLevelStrFlag, err)
	}

	logrus.SetOutput(cmd.OutOrStdout())
	logrus.SetLevel(logLevel)

	return nil
}