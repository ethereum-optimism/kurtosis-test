package backend

import (
	"fmt"
	"os"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"

	bolt "go.etcd.io/bbolt"
)

type KurtestosisValueStore struct {
	RuntimeValueStore *runtime_value_store.RuntimeValueStore
	InterpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore
	StarlarkValueSerde *kurtosis_types.StarlarkValueSerde
}

func CreateValueStores(enclaveDB *enclave_db.EnclaveDB, starlarkValueSerde *kurtosis_types.StarlarkValueSerde) (*runtime_value_store.RuntimeValueStore, *interpretation_time_value_store.InterpretationTimeValueStore, error) {
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(starlarkValueSerde, enclaveDB)
	if err != nil {
		return nil, nil, fmt.Errorf("An error occurred creating the runtime value store: %w", err)
	}

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDB, starlarkValueSerde)
	if err != nil {
		return nil, nil, fmt.Errorf("An error occurred while creating the interpretation time value store: %w", err)
	}

	return runtimeValueStore, interpretationTimeValueStore, nil
}

type TeardownEnclaveDB = func() ()

func CreateEnclaveDB() (*enclave_db.EnclaveDB, TeardownEnclaveDB, error) {
	var err error

	// Create a new temporary enclave database file
	file, err := os.CreateTemp(os.TempDir(), "*.db")
	if err != nil {
		logrus.Errorf("Failed to create temporary database file: %v", err)

		return nil, nil, fmt.Errorf("failed to create temporary database file: %w", err)
	}

	// Now we open the database
	db, err := bolt.Open(file.Name(), 0666, nil)
	if err != nil {
		logrus.Errorf("Failed to open a database in %s: %v", file.Name(), err)

		return nil, nil, fmt.Errorf("failed to open a database in %s: %w", file.Name(), err)
	}

	// Wrap the db
	enclaveDB := &enclave_db.EnclaveDB{
		DB: db,
	}

	// Teardown function that closes the database and removes the temporary file
	// 
	// TODO Technically this can still leave some trash behind if the bolt.Open call fails
	teardown := func () {
		var err error

		err = db.Close()
		if err != nil {
			logrus.Warnf("Failed to close EnclaveDB: %v", err)
		}

		err = os.Remove(file.Name())
		if err != nil {
			logrus.Warnf("Failed to remove EnclaveDB database file %s: %v", file.Name(), err)
		}
	}

	return enclaveDB, teardown, nil
}

func CreateStarlarkValueSerde() *kurtosis_types.StarlarkValueSerde {
	starlarkThread := &starlark.Thread{
		Name:       "kurtosis-starlark-serde-thread",
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := startosis_engine.Predeclared()
	builtins := startosis_engine.KurtosisTypeConstructors()
	for _, builtin := range builtins {
		starlarkEnv[builtin.Name()] = builtin
	}
	return kurtosis_types.NewStarlarkValueSerde(starlarkThread, starlarkEnv)
}
