package kurtosis

import (
	"fmt"
	"kurtestosis/cli/kurtosis/modules"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

func LoadKurtestosisPredeclared(interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore) (starlark.StringDict, error) {
	var err error

	assertPredeclared, err := starlarktest.LoadAssertModule()
	if err != nil {
		return nil, fmt.Errorf("failed to load assert module: %v", err)
	}

	kurtestosisPredeclared, err := modules.LoadKurtestosisModule(interpretationTimeValueStore)
	if err != nil {
		return nil, fmt.Errorf("failed to load assert kurtestosis: %v", err)
	}

	return MergeDicts(assertPredeclared, kurtestosisPredeclared), nil
}

func CreateProcessBuiltins(extraPredeclared starlark.StringDict) startosis_engine.StartosisInterpreterBuiltinsProcessor {
	return func(thread *starlark.Thread, predeclared starlark.StringDict) starlark.StringDict {
		return MergeDicts(predeclared, extraPredeclared)
	}
}

func SetupKurtestosisPredeclared(reporter starlarktest.Reporter) {
	modules.SetBeforeTestFunction(func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) error {
		starlarktest.SetReporter(thread, reporter)

		return nil
	})
}
