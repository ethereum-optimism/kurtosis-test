package backend_test

import (
	"kurtosis-test/cli/kurtosis/backend"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKurtosisTestServiceNetwork(t *testing.T) {

	t.Run("should create unique artifact filenames", func(t *testing.T) {
		var wg sync.WaitGroup
		var artifactNames sync.Map

		serviceNetwork := backend.CreateKurtosisTestServiceNetwork()

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				artifactName, err := serviceNetwork.GetUniqueNameForFileArtifact()
				require.NoError(t, err)

				_, loaded := artifactNames.LoadOrStore(artifactName, true)
				require.False(t, loaded, "Artifact name should be unique")
			}()
		}

		wg.Wait()
	})
}
