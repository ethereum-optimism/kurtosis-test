package commands

import (
	"os"
	"path"
)

var KurtestosisCmdStr = path.Base(os.Args[0])

const (
	KurtestosisDefaultTempDirRoot = ".kurtestosis"

	KurtestosisDefaultTestFilePattern = "**/*_{test,spec}.star"
	
	KurtestosisDefaultTestFunctionPattern = "test_*"
)