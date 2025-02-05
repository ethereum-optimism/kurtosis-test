package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"kurtestosis/cli/runner"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	bolt "go.etcd.io/bbolt"
)

const (
	// CLI Flag names
	logLevelStrFlag = "log-level"
	tempDirRootStrFlag = "temp-dir"
	testFilePatternStrFlag = "test-file-pattern"
	testPatternStrFlag = "test-pattern"

	testEnclaveUUID = "kurtestosis-enclave"
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

	// These are derived from tempDirRootStr
	// 
	// TODO Replace with getters
	repositoriesDirPath string
	githubAuthDirPath string
	tempDirectoriesDirPath string
)


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
	// The PersistentPreRunE hook runs before every descendant command
	// and will setup things like log level
	PersistentPreRunE: setupCLI,
	// The PrerunE will only run before this command and will setup
	// command-specific environment
	PreRunE: setupCommand,
	RunE: run,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&logLevelStr,
		logLevelStrFlag,
		logrus.InfoLevel.String(),
		"Sets the level that the CLI will log at ("+strings.Join(getAllLogLevelStrings(), "|")+")",
	)

	RootCmd.Flags().StringVar(
		&tempDirRootStr,
		tempDirRootStrFlag,
		KurtestosisDefaultTempDirRoot,
		"Directory for kurtosis temporary files",
	)

	RootCmd.Flags().StringVar(
		&testFilePatternStr,
		testFilePatternStrFlag,
		KurtestosisDefaultTestFilePattern,
		"Glob expression to use when looking for starlark test files",
	)

	RootCmd.Flags().StringVar(
		&testPatternStr,
		testPatternStrFlag,
		KurtestosisDefaultTestFunctionPattern,
		"Glob expression to use when looking for test functions",
	)
}

func run(cmd *cobra.Command, args []string) error {
	logrus.Warn("kurtestosis CLI is still work in progress")

	// First we load the project
	projectPath := args[0]
	project, projectErr := runner.LoadProject(args[0])
	if projectErr != nil {
		logrus.Errorf("Failed to load project from %s: %v", projectPath, projectErr)

		return fmt.Errorf("failed to load project from %s: %w", projectPath, projectErr)
	}

	// Let's now get the list of matching test files
	testFiles, testFilesErr := runner.ListMatchingTestFiles(project, testFilePatternStr)
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
	testSuiteSummary := runner.NewTestSuiteSummary(project)

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

	logrus.Errorf("Test suite failed")

	return fmt.Errorf("Test suite failed")
}

func runTestFile(testFile *runner.TestFile) (*runner.TestFileSummary, error) {
	// The summary object will hold the test results for this test file
	testFileSummary := runner.NewTestFileSummary(testFile)

	// First we parse the test file and extract the names of matching test functions
	testFunctions, testFunctionsErr := runner.ListMatchingTests(testFile, testPatternStr)
	if testFunctionsErr != nil {
		logrus.Errorf("Failed to list matching test functions in %s: %v", testFile, testFunctionsErr)
        
		return nil, fmt.Errorf("failed to list matching test functions in %s: %w", testFile, testFunctionsErr)
	}

	// Exit if there are no test suites to run
    if len(testFunctions) == 0 {
        logrus.Warnf("No tests found matching the test pattern %s in %s", testPatternStr, testFile)
        
		return testFileSummary, nil
    }

	logrus.Infof("SUITE %s", testFile)

	// Iterate over the test suites and run them one by one, collecting the test run summaries
	for _, testFunction := range(testFunctions) {
		testFunctionSummary, testFunctionErr := runTestFunction(testFunction)
		if testFunctionErr != nil {
			logrus.Errorf("Failed to run test function %s: %v", testFunction, testFunctionErr)

			return nil, fmt.Errorf("failed to run test function %s: %w", testFunction, testFunctionErr)
		}

		testFileSummary.Append(testFunctionSummary)
	}

	return testFileSummary, nil
}

func runTestFunction(testFunction *runner.TestFunction) (*runner.TestFunctionSummary, error) {
	// The summary object will hold the test results for this test function
	logrus.Debugf("\tRUN %ss", testFunction)

	reporter := runner.NewStarlarktestReporter(testFunction)
	processBuiltins := runner.CreateProcessBuiltins(reporter)

	interpreter, interpreterErr := createInterpreter(testFunction.TestFile.Project, processBuiltins)
	if interpreterErr != nil {
		return nil, interpreterErr
	}

	testSuiteScript := runner.CreateTestSuite(testFunction)

	interpreter.Interpret(
		context.Background(), // context
		testFunction.TestFile.Project.KurotosisYml.PackageName, // packageId
		"", // mainFunctionName
		testFunction.TestFile.Project.KurotosisYml.PackageReplaceOptions, // packageReplaceOptions
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, // relativePathtoMainFile
		testSuiteScript, // serializedStarlark
		startosis_constants.EmptyInputArgs, // serializedJsonParams
		false, // nonBlockingMode 
		enclave_structure.NewEnclaveComponents(), // enclaveComponents
		resolver.NewInstructionsPlanMask(0), // instructionsPlanMask
		image_download_mode.ImageDownloadMode_Missing, // imageDownloadMode
	)

	testFunctionSummary := reporter.Summary()
	if testFunctionSummary.Success() {
		logrus.Infof("\tSUCCESS %s", testFunction)
	} else {
		logrus.Errorf("\tFAIL %s:\n================================================\n%v\n================================================", testFunction, testFunctionSummary.Errors())
	}

	return testFunctionSummary, nil
}

// Creates an enclave
func getEnclaveDBForTest() (*enclave_db.EnclaveDB, error) {
	// Create a new temporary enclave database file
	file, err := os.CreateTemp(os.TempDir(), "*.db")

	// We register a cleanup function
	defer func() {
		err = os.Remove(file.Name())

		if err != nil {
			logrus.Warnf("Failed to remove temporary database file %s: %v", file.Name(), err)
		}
	}()

	// We make sure that the file has been created okay
	if err != nil {
		logrus.Errorf("Failed to create temporary database file: %v", err)

		return nil, fmt.Errorf("failed to create temporary database file: %w", err)
	}

	// Now we open the database
	db, err := bolt.Open(file.Name(), 0666, nil)
	if err != nil {
		logrus.Errorf("Failed to open a database in %s: %v", file.Name(), err)

		return nil, fmt.Errorf("failed to open a database in %s: %w", file.Name(), err)
	}

	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb, nil
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
	// First we configure the log level
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return fmt.Errorf("error parsing the %s CLI argument: %w", logLevelStrFlag, err)
	}

	logrus.SetOutput(cmd.OutOrStdout())
	logrus.SetLevel(logLevel)
	

	return nil
}

// Setup command-specific logic
func setupCommand(cmd *cobra.Command, args []string) error {
	// First we resolve the temporary filesystem paths
	repositoriesDirPath = filepath.Join(tempDirRootStr, "repos")
	githubAuthDirPath = filepath.Join(tempDirRootStr, "auth")
	tempDirectoriesDirPath = filepath.Join(tempDirRootStr, "temp")

	// Then we create the necessary directories
	// 
	// The first one is the path for cloned repositories
	err := os.MkdirAll(repositoriesDirPath, 0700)
	if err != nil {
		return fmt.Errorf("failed to create kurtosis repositories directory: %w", err)
	}

	// Then github credentials
	err = os.MkdirAll(githubAuthDirPath, 0700)
	if err != nil {
		return fmt.Errorf("failed to create kurtosis github auth directory: %w", err)
	}

	// And finally the temporary directories
	err = os.MkdirAll(tempDirectoriesDirPath, 0700)
	if err != nil {
		return fmt.Errorf("failed to create kurtosis temp directories directory: %w", err)
	}

	return nil
}

func createInterpreter(project *runner.KurtestosisProject, processBuiltins startosis_engine.StartosisInterpreterBuiltinsProcessor) (*startosis_engine.StartosisInterpreter, error) {
	// Create test database
	enclaveDb, err := getEnclaveDBForTest()
	if err != nil {
		return nil, err
	}

	// Create package content provider
	githubAuthProvider := git_package_content_provider.NewGitHubPackageAuthProvider(githubAuthDirPath)
	gitPackageContentProvider := git_package_content_provider.NewGitPackageContentProvider(repositoriesDirPath, tempDirectoriesDirPath, githubAuthProvider, enclaveDb)
	wrappedPackageContentProvider := runner.NewLocalProxyPackageContentProvider(project, gitPackageContentProvider)

	serviceNetwork := runner.NewMockServiceNetwork()
	serviceNetwork.EXPECT().GetEnclaveUuid().Maybe().Return(enclave.EnclaveUUID(testEnclaveUUID))
	serviceNetwork.EXPECT().GetApiContainerInfo().Times(1).Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0"),
	)

	starlarkValueSerde := createStarlarkValueSerde()
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(starlarkValueSerde, enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the runtime value store")
	}

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, starlarkValueSerde)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while creating the interpretation time value store")
	}

	return startosis_engine.NewStartosisInterpreterWithBuiltinsProcessor(serviceNetwork, wrappedPackageContentProvider, runtimeValueStore, starlarkValueSerde, "", interpretationTimeValueStore, processBuiltins), nil
}

func createStarlarkValueSerde() *kurtosis_types.StarlarkValueSerde {
	starlarkThread := &starlark.Thread{
		Name:       "starlark-serde-thread",
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := startosis_engine.Predeclared()
	builtins := startosis_engine.KurtosisTypeConstructors()
	for _, builtin := range builtins {
		starlarkEnv[builtin.Name()] = builtin
	}
	return kurtosis_types.NewStarlarkValueSerde(starlarkThread, starlarkEnv)
}