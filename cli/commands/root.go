package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	bolt "go.etcd.io/bbolt"
)

const (
	logLevelStrFlag = "log-level"
	tempDirStrFlag = "temp-dir"

	testEnclaveUUID = "kurtestosis-enclave"
)

// The log level is configurable via the CLI
var logLevelStr string
var defaultLogLevelStr = logrus.InfoLevel.String()

// The directory for temporary files is configurable as well
var tempDirStr string
var defaultTempDirStr = ".kurtestosis"

var repositoriesDirPath string
var githubAuthDirPath string
var tempDirectoriesDirPath string

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

	RootCmd.PersistentFlags().StringVar(
		&tempDirStr,
		tempDirStrFlag,
		defaultTempDirStr,
		"Default directory for kurtosis temporary files",
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

	// Run the test suites
	for _, testSuitePath := range testSuitePaths {
		logrus.Infof("Running test suite from %s", testSuitePath)

        err := runTestSuite(testSuitePath)
        if err != nil {
            logrus.Errorf("Error running test suite %s: %v", testSuitePath, err)
        }
    }

	return nil
}

func runTestSuite(testSuitePath string) error {
	// First we load the test suite script
	testSuiteScript, err := os.ReadFile(testSuitePath)
	if err != nil {
		return err
	}

	// 
	// Then we create the filesystem
	// 

	// The first one is the path for cloned repositories
	err = os.MkdirAll(repositoriesDirPath, 0700)
	if err != nil {
		return err
	}

	// Then github credentials
	err = os.MkdirAll(githubAuthDirPath, 0700)
	if err != nil {
		return err
	}

	// And finally the temporary directories
	err = os.MkdirAll(tempDirectoriesDirPath, 0700)
	if err != nil {
		return err
	}

	// 
	// Now the runtime setup for the interpreter
	// 

	enclaveDb, err := getEnclaveDBForTest()
	if err != nil {
		return err
	}

	githubAuthProvider := git_package_content_provider.NewGitHubPackageAuthProvider(githubAuthDirPath)
	gitPackageContentProvider := git_package_content_provider.NewGitPackageContentProvider(repositoriesDirPath, tempDirectoriesDirPath, githubAuthProvider, enclaveDb)

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	if err != nil {
		return err
	}

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	if err != nil {
		return err
	}

	serviceNetwork := &service_network.MockServiceNetwork{}
	serviceNetwork.EXPECT().GetEnclaveUuid().Maybe().Return(enclave.EnclaveUUID(testEnclaveUUID))

	interpreter := startosis_engine.NewStartosisInterpreter(serviceNetwork, gitPackageContentProvider, runtimeValueStore, nil, "", interpretationTimeValueStore)

	_, _, interpretationError := interpreter.Interpret(
		context.Background(), // context
		startosis_constants.PackageIdPlaceholderForStandaloneScript, // packageId
		"", // FIXME mainFunctionName
		map[string]string{}, // packageReplaceOptions
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, // relativePathtoMainFile
		string(testSuiteScript), // serializedStarlark
		startosis_constants.EmptyInputArgs, // serializedJsonParams
		false, // nonBlockingMode 
		enclave_structure.NewEnclaveComponents(), // enclaveComponents
		resolver.NewInstructionsPlanMask(0), // instructionsPlanMask
		image_download_mode.ImageDownloadMode_Missing, // imageDownloadMode
	)

	if interpretationError != nil {
		logrus.Errorf("Failed to interpret %s: %v", testSuitePath, interpretationError)
		
		return fmt.Errorf("failed to interpret %s: %v", testSuitePath, interpretationError)
	}

	return nil
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

		return nil, fmt.Errorf("Failed to create temporary database file: %w", err)
	}

	// Now we open the database
	db, err := bolt.Open(file.Name(), 0666, nil)
	if err != nil {
		logrus.Errorf("Failed to open a database in %s: %v", file.Name(), err)

		return nil, fmt.Errorf("Failed to open a database in %s: %w", file.Name(), err)
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

	// Now we setup the filesystem
	repositoriesDirPath = filepath.Join(tempDirStr, "repos")
	githubAuthDirPath = filepath.Join(tempDirStr, "auth")
	tempDirectoriesDirPath = filepath.Join(tempDirStr, "temp")

	return nil
}