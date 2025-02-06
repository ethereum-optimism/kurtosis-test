package backend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"
)

const (
	tempDirMode os.FileMode = 0700
)

func CreateLocalGitPackageContentProvider(artifactsPath string, enclaveDB *enclave_db.EnclaveDB) (*git_package_content_provider.GitPackageContentProvider, error) {
	var err error

	// First we resolve the temporary filesystem paths
	repositoriesDirPath := filepath.Join(artifactsPath, "repos")
	tempDirectoriesDirPath := filepath.Join(artifactsPath, "temp")

	// Then we create the necessary directories
	err = createTempDirectory(repositoriesDirPath)
	if err != nil {
		return nil, err
	}

	err = createTempDirectory(tempDirectoriesDirPath)
	if err != nil {
		return nil, err
	}

	// Then we create a package content provider backed by these directories
	//
	// TODO The auth provider is not specified now which means kurtestosis will not be able to use private repos
	return git_package_content_provider.NewGitPackageContentProvider(repositoriesDirPath, tempDirectoriesDirPath, nil, enclaveDB), nil
}

func createTempDirectory(dirPath string) error {
	err := os.MkdirAll(dirPath, tempDirMode)
	if err != nil {
		return fmt.Errorf("failed to create kurtosis temp directories directory: %w", err)
	}

	return nil
}
