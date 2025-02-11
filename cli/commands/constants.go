package commands

import (
	"os"
	"path"
)

var KurtosisTestCmdStr = path.Base(os.Args[0])

const (
	KurtosisTestDefaultTempDirRoot = ".kurtosis-test"

	KurtosisTestDefaultTestFilePattern = "**/*_{test,spec}.star"

	KurtosisTestDefaultTestFunctionPattern = "test_*"
)
