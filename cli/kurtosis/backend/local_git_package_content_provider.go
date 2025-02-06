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
	githubAuthDirPath := filepath.Join(artifactsPath, "auth")
	tempDirectoriesDirPath := filepath.Join(artifactsPath, "temp")

	// Then we create the necessary directories
	err = createTempDirectory(repositoriesDirPath)
	if err != nil {
		return nil, err
	}

	err = createTempDirectory(githubAuthDirPath)
	if err != nil {
		return nil, err
	}

	err = createTempDirectory(tempDirectoriesDirPath)
	if err != nil {
		return nil, err
	}

	// Then we create a package content provider backed by these directories
	//
	// TODO The auth provider is not really doing anything since there is no way of logging in with it
	// so private github kurtosis packages will not work
	githubPackageAuthProvider := git_package_content_provider.NewGitHubPackageAuthProvider(githubAuthDirPath)

	return git_package_content_provider.NewGitPackageContentProvider(repositoriesDirPath, tempDirectoriesDirPath, githubPackageAuthProvider, enclaveDB), nil
}

func createTempDirectory(dirPath string) error {
	err := os.MkdirAll(dirPath, tempDirMode)
	if err != nil {
		return fmt.Errorf("failed to create kurtosis temp directories directory: %w", err)
	}

	return nil
}
