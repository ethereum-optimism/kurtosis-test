package backend

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_identifiers"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

const (
	enclaveUUID = "kurtosis-test-enclave"
)

var (
	apiContainerInfo = service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0")

	// Make sure KurtosisTestServiceNetwork implements service_network.ServiceNetwork
	_ service_network.ServiceNetwork = (*KurtosisTestServiceNetwork)(nil)
)

type KurtosisTestServiceNetwork struct {
	fileArtifactNameCounter atomic.Int64
}

func CreateKurtosisTestServiceNetwork() *KurtosisTestServiceNetwork {
	return &KurtosisTestServiceNetwork{
		fileArtifactNameCounter: atomic.Int64{},
	}
}

func (network *KurtosisTestServiceNetwork) AddService(
	ctx context.Context,
	serviceName service.ServiceName,
	serviceConfig *service.ServiceConfig,
) (
	*service.Service,
	error,
) {
	return nil, unimplemented("AddService")
}

func (network *KurtosisTestServiceNetwork) AddServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceName]*service.ServiceConfig,
	batchSize int,
) (
	map[service.ServiceName]*service.Service,
	map[service.ServiceName]error,
	error,
) {
	return nil, nil, unimplemented("AddServices")
}

func (network *KurtosisTestServiceNetwork) UpdateService(
	ctx context.Context,
	serviceName service.ServiceName,
	updateServiceConfig *service.ServiceConfig,
) (
	*service.Service,
	error,
) {
	return nil, unimplemented("AddServices")
}

func (network *KurtosisTestServiceNetwork) UpdateServices(
	ctx context.Context,
	updateServiceConfigs map[service.ServiceName]*service.ServiceConfig,
	batchSize int,
) (
	map[service.ServiceName]*service.Service,
	map[service.ServiceName]error,
	error,
) {
	return nil, nil, unimplemented("UpdateServices")
}

func (network *KurtosisTestServiceNetwork) RemoveService(ctx context.Context, serviceIdentifier string) (service.ServiceUUID, error) {
	return "", unimplemented("RemoveService")
}

func (network *KurtosisTestServiceNetwork) StartService(ctx context.Context, serviceIdentifier string) error {
	return unimplemented("StartService")
}

func (network *KurtosisTestServiceNetwork) StartServices(
	ctx context.Context,
	serviceIdentifiers []string,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	return nil, nil, unimplemented("StartServices")
}

func (network *KurtosisTestServiceNetwork) StopService(ctx context.Context, serviceIdentifier string) error {
	return unimplemented("StopService")
}

func (network *KurtosisTestServiceNetwork) StopServices(
	ctx context.Context,
	serviceIdentifiers []string,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	return nil, nil, unimplemented("StopServices")
}

func (network *KurtosisTestServiceNetwork) RunExec(ctx context.Context, serviceIdentifier string, userServiceCommand []string) (*exec_result.ExecResult, error) {
	return nil, unimplemented("RunExec")
}

func (network *KurtosisTestServiceNetwork) RunExecs(
	ctx context.Context,
	userServiceCommands map[string][]string,
) (
	map[service.ServiceUUID]*exec_result.ExecResult,
	map[service.ServiceUUID]error,
	error,
) {
	return nil, nil, unimplemented("RunExecs")
}

func (network *KurtosisTestServiceNetwork) HttpRequestService(ctx context.Context, serviceIdentifier string, portId string, method string, contentType string, endpoint string, body string, headers map[string]string) (*http.Response, error) {
	return nil, unimplemented("HttpRequestService")
}

func (network *KurtosisTestServiceNetwork) GetService(ctx context.Context, serviceIdentifier string) (*service.Service, error) {
	return nil, unimplemented("GetService")
}

func (network *KurtosisTestServiceNetwork) GetServices(ctx context.Context) (map[service.ServiceUUID]*service.Service, error) {
	return nil, unimplemented("GetServices")
}

func (network *KurtosisTestServiceNetwork) CopyFilesFromService(ctx context.Context, serviceIdentifier string, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	return "", unimplemented("CopyFilesFromService")
}

func (network *KurtosisTestServiceNetwork) GetServiceNames() (map[service.ServiceName]bool, error) {
	return nil, unimplemented("GetServiceNames")
}

func (network *KurtosisTestServiceNetwork) GetExistingAndHistoricalServiceIdentifiers() (service_identifiers.ServiceIdentifiers, error) {
	return nil, unimplemented("GetExistingAndHistoricalServiceIdentifiers")
}

func (network *KurtosisTestServiceNetwork) ExistServiceRegistration(serviceName service.ServiceName) (bool, error) {
	return false, unimplemented("ExistServiceRegistration")
}

func (network *KurtosisTestServiceNetwork) RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*render_templates.TemplateData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	return "", unimplemented("RenderTemplates")
}

func (network *KurtosisTestServiceNetwork) UploadFilesArtifact(data io.Reader, contentMd5 []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	return "", unimplemented("UploadFilesArtifact")
}

func (network *KurtosisTestServiceNetwork) GetFilesArtifactMd5(artifactName string) (enclave_data_directory.FilesArtifactUUID, []byte, bool, error) {
	return "", nil, false, unimplemented("GetFilesArtifactMd5")
}

func (network *KurtosisTestServiceNetwork) UpdateFilesArtifact(fileArtifactUuid enclave_data_directory.FilesArtifactUUID, updatedContent io.Reader, contentMd5 []byte) error {
	return unimplemented("UpdateFilesArtifact")
}

func (network *KurtosisTestServiceNetwork) GetUniqueNameForFileArtifact() (string, error) {
	artifactIndex := network.fileArtifactNameCounter.Add(1)

	return fmt.Sprintf("file-artifact-%d", artifactIndex), nil
}

func (network *KurtosisTestServiceNetwork) GetApiContainerInfo() *service_network.ApiContainerInfo {
	return apiContainerInfo
}

func (network *KurtosisTestServiceNetwork) GetEnclaveUuid() enclave.EnclaveUUID {
	return enclave.EnclaveUUID(enclaveUUID)
}

func unimplemented(methodName string) error {
	return fmt.Errorf("KurtosisTestServiceNetwork does not support %s method", methodName)
}
