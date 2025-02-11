package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	DebugBuiltinName = "debug"

	DebugBuiltinValueArgName = "value"
)

func NewDebug() *kurtosis_helper.KurtosisHelper {
	return &kurtosis_helper.KurtosisHelper{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: DebugBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              DebugBuiltinValueArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return nil
					},
				},
			},
		},

		Capabilities: &debugCapabilities{},
	}
}

type debugCapabilities struct{}

func (builtin *debugCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	valueArg, err := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, DebugBuiltinValueArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while extracting the value argument for debug builtin")
	}

	logrus.Infof("%s: %s", locatorOfModuleInWhichThisBuiltInIsBeingCalled, valueArg)

	return starlark.None, nil
}
