package backend_test

import (
	"kurtosis-test/cli/kurtosis/backend"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKurtosisTestServiceNetwork(t *testing.T) {
	t.Run("should create unique artifact filenames", func(t *testing.T) {
		serviceNetwork := backend.CreateKurtosisTestServiceNetwork()

		artifactNames := make(map[string]bool)

		for i := 0; i < 10000; i++ {
			artifactName, err := serviceNetwork.GetUniqueNameForFileArtifact()
			require.NoError(t, err)
			require.False(t, artifactNames[artifactName], "Artifact name should be unique")

			artifactNames[artifactName] = true
		}
	})
}
