package core

import (
	"fmt"
	"path/filepath"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/sirupsen/logrus"
)

type KurtestosisProject struct {
	KurotosisYml *enclaves.KurtosisYaml
	Path string
}

func LoadKurtestosisProject(projectPath string) (*KurtestosisProject, error) {
	logrus.Debugf("Loading project from %s", projectPath)

	projectPathAbsolute, projectPathAbsoluteErr := filepath.Abs(projectPath)
	if projectPathAbsoluteErr != nil {
		return nil, fmt.Errorf("failed to determine absolute path to project root %s: %v", projectPath, projectPathAbsoluteErr)
	}

	// At this point we need to load kurtosis.yml and see what's inside
	// 
	// Specifically, we'll need the package name so that we don't make up one ourselves
	// (everything works but stacktraces might be confusing)
	kurtosisYamlFilepath := filepath.Join(projectPathAbsolute, "kurtosis.yml")
	kurtosisYml, kurtosisYmlErr := enclaves.ParseKurtosisYaml(kurtosisYamlFilepath)
	if kurtosisYmlErr != nil {
		logrus.Errorf("Failed to load kurtosis.yml from %s: %v", kurtosisYamlFilepath, kurtosisYmlErr)

		return nil, fmt.Errorf("failed to load kurtosis.yml from %s: %w", kurtosisYamlFilepath, kurtosisYmlErr)
	}
	logrus.Debugf("Loaded kurtosis config from %s", kurtosisYamlFilepath)

	return &KurtestosisProject{
		KurotosisYml: kurtosisYml,
		Path: projectPathAbsolute,
	}, nil
}