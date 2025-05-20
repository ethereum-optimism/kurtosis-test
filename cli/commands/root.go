package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"kurtosis-test/cli/core"
	"kurtosis-test/cli/kurtosis"
	"kurtosis-test/cli/kurtosis/backend"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// CLI Flag names
	logLevelStrFlag        = "log-level"
	tempDirRootStrFlag     = "temp-dir"
	testFilePatternStrFlag = "test-file-pattern"
	testPatternStrFlag     = "test-pattern"
)

// The variables configurable using CLI flags
var (
	// Log level for the CLI
	logLevelStr string

	// Temporary directory in which to store kurtosis' temporary filesystem
	tempDirRootStr string

	// Glob pattern to use when looking for test files
	testFilePatternStr string

	// Glob pattern to use when looking for test functions
	testPatternStr string
)

// RootCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var RootCmd = &cobra.Command{
	Use:   KurtosisTestCmdStr,
	Short: "Kurtosis test runner CLI",
	// Cobra will print usage whenever _any_ error occurs, including ones we throw in Kurtosis
	// This doesn't make sense in 99% of the cases, so just turn them off entirely
	SilenceUsage: true,
	// Cobra prints the errors itself, however, with this flag disabled it give Kurtosis control
	// and allows us to post process the error in the main.go file.
	SilenceErrors: true,
	// The PersistentPreRunE hook runs before every descendant command
	// and will setup things like log level
	PersistentPreRunE: setupCLI,
	RunE:              run,
	Args:              cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&logLevelStr,
		logLevelStrFlag,
		logrus.InfoLevel.String(),
		"Sets the level that the CLI will log at ("+strings.Join(core.ToStringList(logrus.AllLevels), "|")+")",
	)

	RootCmd.Flags().StringVar(
		&tempDirRootStr,
		tempDirRootStrFlag,
		KurtosisTestDefaultTempDirRoot,
		"Directory for kurtosis temporary files",
	)

	RootCmd.Flags().StringVar(
		&testFilePatternStr,
		testFilePatternStrFlag,
		KurtosisTestDefaultTestFilePattern,
		"Glob expression to use when looking for starlark test files",
	)

	RootCmd.Flags().StringVar(
		&testPatternStr,
		testPatternStrFlag,
		KurtosisTestDefaultTestFunctionPattern,
		"Glob expression to use when looking for test functions",
	)
}

func run(cmd *cobra.Command, args []string) error {
	logrus.Warn("kurtosis-test CLI is still work in progress")

	// First we load the project
	projectPath := args[0]
	project, projectErr := core.LoadKurtosisTestProject(args[0])
	if projectErr != nil {
		logrus.Errorf("Failed to load project from %s: %v", projectPath, projectErr)

		return fmt.Errorf("failed to load project from %s: %w", projectPath, projectErr)
	}

	// Let's now get the list of matching test files
	//
	// We need to make sure to ignore the test files in the temporary directory
	testFiles, testFilesErr := core.ListMatchingTestFiles(project, testFilePatternStr, filepath.Join(tempDirRootStr, "**"))
	if testFilesErr != nil {
		logrus.Errorf("Error matching test files in project: %v", testFilesErr)

		return fmt.Errorf("error matching test files in project: %w", testFilesErr)
	}

	// Exit if there are no test suites to run
	if len(testFiles) == 0 {
		logrus.Warn("No test suites found matching the glob pattern")

		return nil
	}

	// The summary of the whole test run
	testSuiteSummary := core.NewTestSuiteSummary(project)

	// Run the test suites
	for _, testFile := range testFiles {
		testFileSummary, err := runTestFile(testFile)
		if err != nil {
			logrus.Errorf("Error running test suite %s: %v", testFile, err)

			return fmt.Errorf("error running test suite %s: %v", testFile, err)
		}

		testSuiteSummary.Append(testFileSummary)
	}

	if testSuiteSummary.Success() {
		return nil
	}

	return fmt.Errorf("test suite failed")
}

func runTestFile(testFile *core.TestFile) (*core.TestFileSummary, error) {
	// The summary object will hold the test results for this test file
	testFileSummary := core.NewTestFileSummary(testFile)

	// First we parse the test file and extract the names of matching test functions
	testFunctions, testFunctionsErr := core.ListMatchingTests(testFile, testPatternStr)
	if testFunctionsErr != nil {
		logrus.Errorf("Failed to list matching test functions in %s: %v", testFile, testFunctionsErr)

		return nil, fmt.Errorf("failed to list matching test functions in %s: %w", testFile, testFunctionsErr)
	}

	// Exit if there are no test suites to run
	if len(testFunctions) == 0 {
		logrus.Debugf("No tests found matching the test pattern %s in %s", testPatternStr, testFile)

		return testFileSummary, nil
	}

	logrus.Infof("SUITE %s", testFile)

	// Iterate over the test suites and run them one by one, collecting the test run summaries
	for _, testFunction := range testFunctions {
		testFunctionSummary, testFunctionErr := runTestFunction(testFunction)
		if testFunctionErr != nil {
			logrus.Errorf("Failed to run test function %s: %v", testFunction, testFunctionErr)

			return nil, fmt.Errorf("failed to run test function %s: %w", testFunction, testFunctionErr)
		}

		testFileSummary.Append(testFunctionSummary)
	}

	return testFileSummary, nil
}

func runTestFunction(testFunction *core.TestFunction) (*core.TestFunctionSummary, error) {
	var err error

	// The summary object will hold the test results for this test function
	logrus.Debugf("\tRUN %ss", testFunction)

	// Let's make a database first
	enclaveDB, teardownEnclaveDB, err := backend.CreateEnclaveDB()
	if err != nil {
		return nil, fmt.Errorf("failed to create EnclaveDB: %w", err)
	}

	// We want to tear the database down once it's all over
	defer teardownEnclaveDB()

	// Package content providers
	localGitPackageContentProvider, err := backend.CreateLocalGitPackageContentProvider(tempDirRootStr, enclaveDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create local git package content provider: %w", err)
	}
	localProxyPackageContentProvider := backend.CreateLocalProxyPackageContentProvider(testFunction.TestFile.Project, localGitPackageContentProvider)

	// Now we create the value storage that holds all the starlark values
	starlarkValueSerde := backend.CreateStarlarkValueSerde()
	runtimeValueStore, interpretationTimeValueStore, err := backend.CreateValueStores(enclaveDB, starlarkValueSerde)
	if err != nil {
		return nil, fmt.Errorf("failed to create kurtosis value stores: %w", err)
	}

	// We load all the kurtosis-test-specific predeclared starlark builtins
	predeclared, err := kurtosis.LoadKurtosisTestPredeclared(interpretationTimeValueStore)
	if err != nil {
		return nil, err
	}

	// And we create a processor function that merges them with kurtosis predeclared builtins
	processBuiltins := kurtosis.CreateProcessBuiltins(predeclared)

	// We setup a test reporter
	//
	// Besides collecting and formatting the test output (mostly TBD),
	// a reporter is required for correct functioning of the starlarktest assert module
	reporter := core.NewTestReporter(testFunction)
	kurtosis.SetupKurtosisTestPredeclared(reporter)

	// Service network (99% mock)
	serviceNetwork := backend.CreateKurtosisTestServiceNetwork()

	// And finally an interpreter
	interpreter, err := backend.CreateInterpreter(
		localProxyPackageContentProvider, // packageContentProvider
		starlarkValueSerde,               // starlarkValueSerde
		runtimeValueStore,                // runtimeValueStore
		interpretationTimeValueStore,     // interpretationTimeValueStore
		processBuiltins,                  // processBuiltins
		serviceNetwork,                   // serviceNetwork
	)
	if err != nil {
		return nil, err
	}

	testSuiteScript, mainFunctionName, inputArgs := kurtosis.WrapTestFunction(testFunction)

	_, _, interpretationErr := interpreter.Interpret(
		context.Background(), // context
		testFunction.TestFile.Project.KurotosisYml.PackageName, // packageId
		mainFunctionName, // mainFunctionName
		testFunction.TestFile.Project.KurotosisYml.PackageReplaceOptions, // packageReplaceOptions
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,  // relativePathtoMainFile
		testSuiteScript,                          // serializedStarlark
		inputArgs,                                // serializedJsonParams
		false,                                    // nonBlockingMode
		enclave_structure.NewEnclaveComponents(), // enclaveComponents
		resolver.NewInstructionsPlanMask(0),      // instructionsPlanMask
		image_download_mode.ImageDownloadMode_Missing, // imageDownloadMode
	)

	// We add any interpretation errors to the summary
	if interpretationErr != nil {
		reporter.Error(interpretationErr)
	}

	// FIXME The reporter should be doing all the lifting when it comes to logging and formatting
	// the test output, at the moment it's kinda ready for that but not utitlized at all

	testFunctionSummary := reporter.Summary()

	if testFunctionSummary.Success() {
		logrus.Infof("\tSUCCESS %s", testFunction)
	} else {
		errorsList := core.ToStringList(testFunctionSummary.Errors())
		errorsString := strings.ReplaceAll(strings.Join(errorsList, "\n\n"), "\\n", "\n")
		errorsSeparator := "================================================"

		logrus.Errorf("\tFAIL %s:\n%s\n%v\n%s", testFunction, errorsSeparator, errorsString, errorsSeparator)
	}

	return testFunctionSummary, nil
}

// Setup function to run before any command execution
func setupCLI(cmd *cobra.Command, args []string) error {
	// First we configure the log level
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return fmt.Errorf("error parsing the %s CLI argument: %w", logLevelStrFlag, err)
	}

	logrus.SetOutput(cmd.OutOrStdout())
	logrus.SetLevel(logLevel)

	return nil
}
