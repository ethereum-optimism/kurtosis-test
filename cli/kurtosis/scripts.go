package kurtosis

import (
	"fmt"
	"kurtestosis/cli/core"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
)

// Creates a wrapper script that executes testFunction
// using the kurtestosis starlark module
//
// This module sets up necessary starlark runtime (especially for the assert module)
func WrapTestFunction(testFunction *core.TestFunction) (starlark string, mainFunctionName string, jsonInputArgs string) {
	return fmt.Sprintf(`
sut = import_module("/%s")

def run(plan):
	kurtestosis.test(plan, sut, "%s")
`, testFunction.TestFile.Path, testFunction.Name), "run", startosis_constants.EmptyInputArgs
}
