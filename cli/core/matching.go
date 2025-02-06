package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"go.starlark.net/syntax"
	"gopkg.in/godo.v2/glob"
)

type TestFile struct {
	Project *KurtestosisProject
	Path string
}

func (testFile *TestFile) String() string {
	return testFile.Path
}

type TestFunction struct {
	TestFile *TestFile
	Name string
}

func (testFunction *TestFunction) String() string {
	return fmt.Sprintf("%s:%s", testFunction.TestFile, testFunction.Name)
}

func ListMatchingTestFiles(project *KurtestosisProject, testFilePattern string) ([]*TestFile, error) {
	// The testFilePattern is expected to be a relative path from the project root
	// so we first need to make sure it will only match inside the project root
	testFilePatternAbsoute := filepath.Join(project.Path, testFilePattern)
	logrus.Debugf("Looking for test files matching %s", testFilePatternAbsoute)
	
	// We look for the matching files
	testSuiteFileAssets, _, globErr := glob.Glob([]string{testFilePatternAbsoute})
    if globErr != nil {
        return nil, fmt.Errorf("error expanding glob pattern: %w", globErr)
    }

	// Now we turn the globbing results into an array of TestFile objects
	testFiles := []*TestFile{}
	for _, testSuiteFileAsset := range testSuiteFileAssets {
		testFilePath := testSuiteFileAsset.Path

		// We let the user know if the match is a directory and continue
		if testSuiteFileAsset.IsDir() {
			logrus.Debugf("Skipping matched test file %s because it's a directory", testFilePath); continue
		}

		logrus.Debugf("Matched test file %s", testFilePath)

		testFilePathRel, testFilePathRelErr := filepath.Rel(project.Path, testFilePath)
		if testFilePathRelErr != nil {
			logrus.Warnf("Failed to determine relative path of test file %s from project root %s: %v", testFilePath, project.Path, testFilePathRelErr); continue
		}

		testFiles = append(testFiles, &TestFile{
			Project: project,
			Path: testFilePathRel,
		})
    }

	logrus.Debugf("Matched %d test files", len(testFiles))

	return testFiles, nil
}

func ListMatchingTests(testFile *TestFile, testPattern string) ([]*TestFunction, error) {
	// First we read the contents of the test file
	testFilePath := filepath.Join(testFile.Project.Path, testFile.Path)
	testScript, testScriptErr := os.ReadFile(testFilePath)
	if testScriptErr != nil {
		logrus.Errorf("Failed to read test suite %s: %v", testFilePath, testScriptErr)

		return nil, fmt.Errorf("failed to read test suite %s: %w", testFilePath, testScriptErr)
	}

	// Now we parse the script contents into a starlark syntax tree
	testParseTree, testParseTreeErr := syntax.Parse(testFile.Path, testScript, 0)
	if testParseTreeErr != nil {
		logrus.Errorf("Failed to parse test suite %s: %v", testFile.Path, testParseTreeErr)

		return nil, fmt.Errorf("failed to parse test suite %s: %w", testFile.Path, testParseTreeErr)
	}

	// Now we walk the starlark tree looking for top-level def statements
	defStmts := []*syntax.DefStmt{}
	syntax.Walk(testParseTree, func(node syntax.Node) bool {
		// If we are looking at the top-level file node, we just continue
		// since we need to look inside the file
		if _, ok := node.(*syntax.File); ok {
            return true
        }

		// If we found a def, we remember it and don't traverse deeper - we are only looking for top-level def statements
		if fn, ok := node.(*syntax.DefStmt); ok {
			defStmts = append(defStmts, fn)

			return false
        }

		// For any other nodes we'll not traverse further since we are only looking for top-level test methods
		return false
	})

	// We turn the test pattern glob into a regexp
	testRegexp := glob.Globexp(testPattern)

	// Now let's filter out the test functions
	testFunctions := []*TestFunction{}
	for _, defStmt := range(defStmts) {
		// First we check that the test function's name matches the test pattern
		if !testRegexp.MatchString(defStmt.Name.Name) {
			logrus.Debugf("Function %s from %s does not match test pattern %s, skipping", defStmt.Name.Name, testFile.Path, testPattern); continue
		}

		// Now we make sure that the function accepts one parameter
		numParams := len(defStmt.Params)
		if numParams != 1 {
			logrus.Warnf("Function %s from %s matches test pattern %s but accepts %d params instead of 1. Test functions should accept only plan param", defStmt.Name.Name, testFile.Path, testPattern, numParams); continue
		}

		testFunctions = append(testFunctions, &TestFunction{
			TestFile: testFile,
			Name: defStmt.Name.Name,
		})
	}

	return testFunctions, nil
}