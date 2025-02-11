package backend

import (
	"os"
	"strings"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/sirupsen/logrus"

	"kurtosis-test/cli/core"
)

// LocalProxyPackageContentProvider wraps an existing package content provider
// to resolve local packages without accessing github
type LocalProxyPackageContentProvider struct {
	startosis_packages.PackageContentProvider
	Project *core.KurtosisTestProject
}

func CreateLocalProxyPackageContentProvider(project *core.KurtosisTestProject, packageContentProvider startosis_packages.PackageContentProvider) *LocalProxyPackageContentProvider {
	return &LocalProxyPackageContentProvider{
		PackageContentProvider: packageContentProvider,
		Project:                project,
	}
}

func (provider *LocalProxyPackageContentProvider) GetModuleContents(absoluteModuleLocator *startosis_packages.PackageAbsoluteLocator) (string, *startosis_errors.InterpretationError) {
	// This provider will check whether the git URL matches our local project and if so,
	// will substitute remote git queries for local ones
	gitUrl := absoluteModuleLocator.GetGitURL()
	packageName := provider.Project.KurotosisYml.PackageName
	packageRoot := provider.Project.Path

	// Let's see if the requested module comes from the local package
	if strings.HasPrefix(gitUrl, packageName) {
		// If so, we'll replace the git URL with a local path
		localName := strings.Replace(gitUrl, packageName, packageRoot, 1)
		logrus.Debugf("Loading module content for %s from %s", gitUrl, localName)

		// And load the contents from disk
		content, contentErr := os.ReadFile(localName)
		if contentErr != nil {
			logrus.Errorf("Failed to load module content from %s: %v", localName, contentErr)

			return "", startosis_errors.NewInterpretationError("Failed to load module content from %s: %v", localName, contentErr)
		}

		return string(content), nil
	}

	// Any non-local queries are proxied to the wrapped PackageContentProvider
	return provider.PackageContentProvider.GetModuleContents(absoluteModuleLocator)
}
