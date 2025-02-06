package starlark

import (
	_ "embed"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	once   sync.Once
	kurtestosis starlark.StringDict
	//go:embed kurtestosis.star
	kurtestosisFileSrc string
	kurtestosisErr     error
	beforeTest KurtestosisHook
	afterTest KurtestosisHook
)

// Type of a function that can be registered as a before/after hook
type KurtestosisHook func (thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) error

// LoadKurtestosisModule loads the assert module.
// It is concurrency-safe and idempotent.
func LoadKurtestosisModule() (starlark.StringDict, error) {
	once.Do(func() {
		predeclared := starlark.StringDict{
			"module":   starlark.NewBuiltin("module", starlarkstruct.MakeModule),
			"__before_test__":    starlark.NewBuiltin("__before_test__", run_before_test),
			"__after_test__":    starlark.NewBuiltin("__after_test__", run_after_test),
		}
		thread := new(starlark.Thread)
		kurtestosis, kurtestosisErr = starlark.ExecFile(thread, "kurtestosis.star", kurtestosisFileSrc, predeclared)
	})

	return kurtestosis, kurtestosisErr
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

func run_before_test(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple)  (starlark.Value, error) {
	if beforeTest == nil {
		return starlark.None.Truth(), nil
	}

	return starlark.None.Truth(), beforeTest(thread, builtin, args, kwargs)
}

func run_after_test(thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple)  (starlark.Value, error) {
	if afterTest == nil {
		return starlark.None.Truth(), nil
	}

	return starlark.None.Truth(), afterTest(thread, builtin, args, kwargs)
}