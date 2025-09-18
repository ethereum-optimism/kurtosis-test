package backend

import (
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
)

func CreateInterpreter(
	packageContentProvider startosis_packages.PackageContentProvider,
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore,
	processBuiltins startosis_engine.StartosisInterpreterBuiltinsProcessor,
	serviceNetwork service_network.ServiceNetwork,
	kurtosisBackendType args.KurtosisBackendType,
) (*startosis_engine.StartosisInterpreter, error) {
	return startosis_engine.NewStartosisInterpreterWithBuiltinsProcessor(
		serviceNetwork,
		packageContentProvider,
		runtimeValueStore,
		starlarkValueSerde,
		"", // No enviornment variables
		interpretationTimeValueStore,
		processBuiltins,
		kurtosisBackendType,
	), nil
}
