package builtins

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	kurtosis_port_spec "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	GetServiceConfigBuiltinName = "get_service_config"

	ServiceNameArgName = "service_name"

	bytesInMegaBytes int64 = 1024 * 1024
)

func NewGetServiceConfig(
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore,
) *kurtosis_helper.KurtosisHelper {
	return &kurtosis_helper.KurtosisHelper{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: GetServiceConfigBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
			},
		},

		Capabilities: &getServiceConfigCapabilities{
			interpretationTimeValueStore: interpretationTimeValueStore,
		},
	}
}

type getServiceConfigCapabilities struct {
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore
}

func (builtin *getServiceConfigCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	var err error

	// First we validate the arguments
	serviceNameArgValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, explicitInterpretationError(err)
	}

	// We convert the service name argument to the required format
	serviceNameStr := serviceNameArgValue.GoString()
	serviceName := service.ServiceName(serviceNameStr)

	// Now we get the service config
	serviceConfig, err := builtin.interpretationTimeValueStore.GetServiceConfig(serviceName)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("Failed to get service config for service %s: %v", serviceNameStr, err)
	}

	// And convert the kurtosis type to starlark struct
	serviceConfigStarlark, interpretationErr := toStarlarkServiceConfig(serviceNameStr, serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return serviceConfigStarlark, interpretationErr
}

func explicitInterpretationError(err error) *startosis_errors.InterpretationError {
	return startosis_errors.WrapWithInterpretationError(
		err,
		"Unable to parse arguments of command '%s'. It should be a non empty string containing a name of a kurtosis service",
		GetServiceConfigBuiltinName)
}

// Converts a kurtosis object into a starlark ServiceConfig struct,
// opposite of how ToKurtosisType() method works
//
// TODO This feels like it should be included in the kurtosis core
// TODO Some of the fields are still not mapped here due to the added complexity
func toStarlarkServiceConfig(serviceName string, serviceConfig *service.ServiceConfig) (*service_config.ServiceConfig, *startosis_errors.InterpretationError) {
	var err *startosis_errors.InterpretationError

	ports, err := portSpecMapToStarlarkDict(serviceName, serviceConfig.GetPrivatePorts())
	if err != nil {
		return nil, err
	}

	publicPorts, err := portSpecMapToStarlarkDict(serviceName, serviceConfig.GetPublicPorts())
	if err != nil {
		return nil, err
	}

	envVars, err := stringMapToStarlarkDict(serviceConfig.GetEnvVars())
	if err != nil {
		return nil, err
	}

	labels, err := stringMapToStarlarkDict(serviceConfig.GetLabels())
	if err != nil {
		return nil, err
	}

	nodeSelectors, err := stringMapToStarlarkDict(serviceConfig.GetNodeSelectors())
	if err != nil {
		return nil, err
	}

	filesToBeMoved, err := stringMapToStarlarkDict(serviceConfig.GetFilesToBeMoved())
	if err != nil {
		return nil, err
	}

	files, err := filesToStarlarkValue(serviceConfig.GetFilesArtifactsExpansion(), serviceConfig.GetPersistentDirectories())
	if err != nil {
		return nil, err
	}

	args := []starlark.Value{
		starlark.String(serviceConfig.GetContainerImageName()), // image
		ports,       // ports
		publicPorts, // publicPorts
		files,       // files
		stringArrayToStarlarkList(serviceConfig.GetEntrypointArgs()), // entrypointArgs
		stringArrayToStarlarkList(serviceConfig.GetCmdArgs()),        // cmdArgs
		envVars, // env_vars
		starlark.String(serviceConfig.GetPrivateIPAddrPlaceholder()),         // private_ip_address_placeholder
		starlark.MakeUint64(serviceConfig.GetCPUAllocationMillicpus()),       // DEPRECATED cpu_allocation
		starlark.MakeUint64(serviceConfig.GetMemoryAllocationMegabytes()),    // DEPRECATED memory_allocation
		starlark.MakeUint64(serviceConfig.GetCPUAllocationMillicpus()),       // max_cpu
		starlark.MakeUint64(serviceConfig.GetMinCPUAllocationMillicpus()),    // min_cpu
		starlark.MakeUint64(serviceConfig.GetMemoryAllocationMegabytes()),    // max_memory
		starlark.MakeUint64(serviceConfig.GetMinMemoryAllocationMegabytes()), // min_memory
		starlark.None,  // TODO ready_conditions - these are not accessible it seems
		labels,         // labels
		starlark.None,  // TODO user
		starlark.None,  // TODO tolerations
		nodeSelectors,  // node_selectors
		filesToBeMoved, // files_to_be_moved
		starlark.Bool(serviceConfig.GetTiniEnabled()), // tini_enabled
	}

	argumentDefinitions := service_config.NewServiceConfigType().Arguments
	argumentValuesSet := builtin_argument.NewArgumentValuesSet(argumentDefinitions, args)
	kurtosisDefaultValue, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(service_config.ServiceConfigTypeName, argumentValuesSet)
	if err != nil {
		return nil, err
	}

	return &service_config.ServiceConfig{
		KurtosisValueTypeDefault: kurtosisDefaultValue,
	}, nil
}

func stringArrayToStarlarkList(input []string) (output *starlark.List) {
	values := []starlark.Value{}
	for _, v := range input {
		values = append(values, starlark.String(v))
	}

	return starlark.NewList(values)
}

func stringMapToStarlarkDict(input map[string]string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	dict := starlark.NewDict(len(input))
	for k, v := range input {
		err := dict.SetKey(starlark.String(k), starlark.String(v))
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "failed to set key")
		}
	}

	return dict, nil
}

func portSpecMapToStarlarkDict(serviceName string, input map[string]*port_spec.PortSpec) (*starlark.Dict, *startosis_errors.InterpretationError) {
	dict := starlark.NewDict(len(input))
	for k, v := range input {
		mapped, err := portSpecMapToStarlarkValue(serviceName, v)
		if err != nil {
			return nil, err
		}

		setErr := dict.SetKey(starlark.String(k), mapped.Struct)
		if setErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "failed to set key")
		}
	}

	return dict, nil
}

func portSpecMapToStarlarkValue(serviceName string, portSpec *port_spec.PortSpec) (*kurtosis_port_spec.PortSpec, *startosis_errors.InterpretationError) {
	var maybeWaitTimeout string
	if portSpec.GetWait() != nil {
		maybeWaitTimeout = portSpec.GetWait().GetTimeout().String()
	}

	kurtosisPortSpec, err := kurtosis_port_spec.CreatePortSpecUsingGoValues(
		serviceName,
		portSpec.GetNumber(),
		portSpec.GetTransportProtocol(),
		portSpec.GetMaybeApplicationProtocol(),
		maybeWaitTimeout,
		portSpec.GetUrl(),
	)

	return kurtosisPortSpec, err
}

func persistenDirectoryToStarlarkValue(persistenDirectory service_directory.PersistentDirectory) (*directory.Directory, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		nil, // artifact_names
		starlark.String(persistenDirectory.PersistentKey),                     // persistent_key
		starlark.MakeInt64(int64(persistenDirectory.Size) / bytesInMegaBytes), // size
	}

	return directoryArgsToDirectory(args)
}

func directoryToStarlarkValue(fileArtifactNames []string) (*directory.Directory, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		stringArrayToStarlarkList(fileArtifactNames), // artifact_names
		nil, // persistent_key
		nil, // size
	}

	return directoryArgsToDirectory(args)
}

func directoryArgsToDirectory(args []starlark.Value) (*directory.Directory, *startosis_errors.InterpretationError) {
	argumentDefinitions := directory.NewDirectoryType().KurtosisBaseBuiltin.Arguments
	argumentValuesSet := builtin_argument.NewArgumentValuesSet(argumentDefinitions, args)
	kurtosisDefaultValue, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(directory.DirectoryTypeName, argumentValuesSet)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &directory.Directory{
		KurtosisValueTypeDefault: kurtosisDefaultValue,
	}, nil
}

func filesToStarlarkValue(filesArtifactExpansion *service_directory.FilesArtifactsExpansion, persistentDirectories *service_directory.PersistentDirectories) (*starlark.Dict, *startosis_errors.InterpretationError) {
	filesStarlark := starlark.NewDict(0)

	if filesArtifactExpansion != nil {
		for key, fileArtifactNames := range filesArtifactExpansion.ServiceDirpathsToArtifactIdentifiers {
			directoryStarlark, err := directoryToStarlarkValue(fileArtifactNames)
			if err != nil {
				return nil, err
			}

			setKeyErr := filesStarlark.SetKey(starlark.String(key), directoryStarlark)
			if setKeyErr != nil {
				return nil, startosis_errors.WrapWithInterpretationError(setKeyErr, "failed to set key")
			}
		}
	}

	if persistentDirectories != nil {
		for key, persistentDirectory := range persistentDirectories.ServiceDirpathToPersistentDirectory {
			persistentDirectoryStarlark, err := persistenDirectoryToStarlarkValue(persistentDirectory)
			if err != nil {
				return nil, err
			}

			setKeyErr := filesStarlark.SetKey(starlark.String(key), persistentDirectoryStarlark)
			if setKeyErr != nil {
				return nil, startosis_errors.WrapWithInterpretationError(setKeyErr, "failed to set key")
			}
		}
	}

	return filesStarlark, nil
}
