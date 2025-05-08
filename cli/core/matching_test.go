package core_test

import (
	"kurtosis-test/cli/commands"
	"kurtosis-test/cli/core"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListMatchingTestFilesIgnorePattern(t *testing.T) {
	t.Run("should not ignore any files if called with empty ignored pattern", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)

		testProjectPath := filepath.Join(cwd, "testdata", "test_project")
		testProject := &core.KurtosisTestProject{
			Path: testProjectPath,
		}

		testFiles, err := core.ListMatchingTestFiles(testProject, commands.KurtosisTestDefaultTestFilePattern, "")
		require.NoError(t, err)
		require.Equal(t, []*core.TestFile{{Project: testProject, Path: "test/nonignored_test.star"}, {Project: testProject, Path: "ignored/ignored_test.star"}}, testFiles)
	})

	t.Run("should ignore matching files if called with relative ignored pattern", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)

		testProjectPath := filepath.Join(cwd, "testdata", "test_project")
		testProject := &core.KurtosisTestProject{
			Path: testProjectPath,
		}

		ignoredTestFilePattern := filepath.Join("ignored", "**")

		testFiles, err := core.ListMatchingTestFiles(testProject, commands.KurtosisTestDefaultTestFilePattern, ignoredTestFilePattern)
		require.NoError(t, err)
		require.Equal(t, []*core.TestFile{{Project: testProject, Path: "test/nonignored_test.star"}}, testFiles)
	})

	t.Run("should ignore matching files if called with absolute ignored pattern", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)

		testProjectPath := filepath.Join(cwd, "testdata", "test_project")
		testProject := &core.KurtosisTestProject{
			Path: testProjectPath,
		}

		ignoredTestFilePattern := filepath.Join(testProjectPath, "ignored", "**")

		testFiles, err := core.ListMatchingTestFiles(testProject, commands.KurtosisTestDefaultTestFilePattern, ignoredTestFilePattern)
		require.NoError(t, err)
		require.Equal(t, []*core.TestFile{{Project: testProject, Path: "test/nonignored_test.star"}}, testFiles)
	})
}
