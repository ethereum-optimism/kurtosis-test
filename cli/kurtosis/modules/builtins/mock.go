package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	MockBuiltinName       = "mock"
	MockBuiltinStructName = "mock"

	MockBuiltinTargetArgName     = "target"
	MockBuiltinMethodNameArgName = "method_name"
)

type starlarkCallable interface {
	CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)
}

func NewMock() *kurtosis_helper.KurtosisHelper {
	return &kurtosis_helper.KurtosisHelper{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: MockBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              MockBuiltinTargetArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// TODO
						return nil
					},
				},
				{
					Name:              MockBuiltinMethodNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
			},
		},

		Capabilities: &mockCapabilities{},
	}
}

type mockCapabilities struct{}

func (builtin *mockCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	// Since we are trying to extract a module passed as an argument, we'd always get a warning
	// about the fact that module cannot be copied
	//
	// It's something we know so let's just suppress the warning
	logLevel := logrus.GetLevel()
	logrus.SetLevel(logrus.ErrorLevel)
	targetArg, _ := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, MockBuiltinTargetArgName)
	logrus.SetLevel(logLevel)

	target, ok := targetArg.(*starlarkstruct.Module)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("mock: only module mocks are possible at the moment. %v is not a module", targetArg)
	}

	// We need to make sure that the module has a property called methodName
	methodNameArg, _ := builtin_argument.ExtractArgumentValue[starlark.String](arguments, MockBuiltinMethodNameArgName)
	methodName := methodNameArg.GoString()
	if !target.Members.Has(methodName) {
		return nil, startosis_errors.NewInterpretationError("mock: module %v doesn't have a property called %s", targetArg, methodName)
	}

	// Now we have to make sure that the property under methodName is a callable function
	var targetMethod starlarkCallable
	targetValue := target.Members[methodName]
	targetMethod, ok = targetValue.(*starlark.Builtin)
	if !ok {
		targetMethod, ok = targetValue.(*starlark.Function)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("mock: property %s of module %v is not a function, it's %v", methodName, targetArg, targetValue)
		}
	}

	mock, mockedMethod := createMock(targetMethod)

	// TODO Add a restore() functionality to the mock
	target.Members[methodName] = mockedMethod

	return mock, nil
}

func createMock(originalMethod starlarkCallable) (mock *starlarkstruct.Struct, mockedMethod *starlark.Builtin) {
	// This will hold an array of structs representing each call to the mocked method
	calls := []starlark.Value{}

	// return values can be mocked, in which case this will be set to a non-nil value
	var mockReturnValue starlark.Value

	mockedMethod = starlark.NewBuiltin(MockBuiltinName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// First we turn the mocked method args/kwargs into starlark values
		callArgs := starlark.NewList(args)
		callKwargs, err := kwargsToDict(kwargs)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "mock: failed to convert kwargs to dict")
		}

		// Now we handle the return value
		var returnValue starlark.Value

		if mockReturnValue != nil {
			// If it's mocked then we just return the mocked value
			returnValue = mockReturnValue
		} else {
			// If it's not we need to call the original method
			returnValue, err = originalMethod.CallInternal(thread, args, kwargs)
			if err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err, "failed to call original method")
			}
		}

		// We create a struct representing this call
		calls = append(calls, starlarkstruct.FromStringDict(starlarkstruct.Default, map[string]starlark.Value{
			"args":         callArgs,
			"kwargs":       callKwargs,
			"return_value": returnValue,
		}))

		return returnValue, nil
	})

	mockMembers := map[string]starlark.Value{
		"original": originalMethod.(starlark.Value),
		"calls": starlark.NewBuiltin("calls", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			return starlark.NewList(calls), nil
		}),
		"mock_return_value": starlark.NewBuiltin("mock_return_value", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			if args.Len() == 0 {
				mockReturnValue = nil
			} else {
				mockReturnValue = args.Index(0)
			}

			return mock, nil
		}),
	}

	mock = starlarkstruct.FromStringDict(starlark.String(MockBuiltinStructName), mockMembers)

	return mock, mockedMethod
}

func kwargsToDict(kwargs []starlark.Tuple) (*starlark.Dict, error) {
	dict := starlark.NewDict(len(kwargs))
	for _, kwarg := range kwargs {
		key, value := kwarg[0], kwarg[1]

		err := dict.SetKey(key, value)
		if err != nil {
			return nil, err
		}
	}

	return dict, nil
}
