package runner

import (
	"fmt"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

const (
	reporterStarlarkName = "setup_reporter"
)

func CreateTestSuite(testFunction *TestFunction) string {
	return fmt.Sprintf(`
sut = import_module("/%s")

def run(plan):
	%s()

	sut.%s(plan)
`, testFunction.TestFile.Path, reporterStarlarkName, testFunction.Name)
}

func CreateProcessBuiltins(reporter starlarktest.Reporter) startosis_engine.StartosisInterpreterBuiltinsProcessor {
	return func(thread *starlark.Thread, predeclared starlark.StringDict) starlark.StringDict {
		assertPredeclared, assertModuleErr := starlarktest.LoadAssertModule()
		if assertModuleErr != nil {
			thread.Cancel(fmt.Sprintf("Failed to load assert module: %v", assertModuleErr))
		}

		for k,v := range(assertPredeclared) {
			predeclared[k] = v
		}

		predeclared[reporterStarlarkName] = starlark.NewBuiltin(reporterStarlarkName, func (thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			starlarktest.SetReporter(thread, reporter)

			return starlark.None.Truth(), nil
		})

		return predeclared
	}
}