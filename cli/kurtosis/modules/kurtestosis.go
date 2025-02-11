package modules

import (
	_ "embed"
	"kurtestosis/cli/kurtosis/modules/builtins"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	//go:embed kurtestosis.star
	kurtestosisFileSrc string
	beforeTest         KurtestosisHook
	afterTest          KurtestosisHook
)

// Type of a function that can be registered as a before/after hook
type KurtestosisHook func(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) error

// LoadKurtestosisModule loads the kurtestosis module.
func LoadKurtestosisModule(interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore) (starlark.StringDict, error) {
	predeclared := starlark.StringDict{
		"module":                             starlark.NewBuiltin("module", starlarkstruct.MakeModule),
		"__before_test__":                    starlark.NewBuiltin("__before_test__", runBeforeTest),
		"__after_test__":                     starlark.NewBuiltin("__after_test__", runAfterTest),
		builtins.GetServiceConfigBuiltinName: starlark.NewBuiltin(builtins.GetServiceConfigBuiltinName, builtins.NewGetServiceConfig(interpretationTimeValueStore).CreateBuiltin()),
		builtins.DebugBuiltinName:            starlark.NewBuiltin(builtins.DebugBuiltinName, builtins.NewDebug().CreateBuiltin()),
	}
	thread := new(starlark.Thread)

	return starlark.ExecFile(thread, "kurtestosis.star", kurtestosisFileSrc, predeclared)
}

// Sets the beforeTest hook, overriding the previous value
//
// beforeTest hook gets executed before every kurtosis test and gets passed
// information about the starlark thread along with context information about the starlark test
func SetBeforeTestFunction(fn KurtestosisHook) {
	beforeTest = fn
}

// Sets the afterTest hook, overriding the previous value
//
// afterTest hook gets executed after every kurtosis test and gets passed
// information about the starlark thread along with context information about the starlark test
func SetAfterTestFunction(fn KurtestosisHook) {
	afterTest = fn
}

func runBeforeTest(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if beforeTest == nil {
		return starlark.None.Truth(), nil
	}

	return starlark.None.Truth(), beforeTest(thread, builtin, args, kwargs)
}

func runAfterTest(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if afterTest == nil {
		return starlark.None.Truth(), nil
	}

	return starlark.None.Truth(), afterTest(thread, builtin, args, kwargs)
}
