service_name = "my-service"
image_name = "dependency"

def test_get_service_config(plan):
    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = image_name,
            labels = {
                "my-label": "my-label-value"
            },
            node_selectors = {
                "select": "node"
            },
            entrypoint = [
                "hello"
            ],
            cmd = [
                "say"
            ],
            env_vars = {
                "OP": "YES"
            },
            min_cpu = 14,
            max_cpu = 15,
            min_memory = 9,
            max_memory = 20,
            tini_enabled = True
        ),
    )

    service_config = kurtosistest.get_service_config(service_name = service_name)

    assert.ne(service_config, None)
    assert.eq(service_config.image, image_name)
    assert.eq(service_config.private_ip_address_placeholder, "KURTOSIS_IP_ADDR_PLACEHOLDER")
    assert.eq(service_config.labels, { "my-label": "my-label-value" })
    assert.eq(service_config.ports, {})
    assert.eq(service_config.public_ports, {})
    assert.eq(service_config.node_selectors, { "select": "node" })
    assert.eq(service_config.entrypoint, ["hello"])
    assert.eq(service_config.env_vars, { "OP": "YES" })
    assert.eq(service_config.cmd, ["say"])
    assert.eq(service_config.min_cpu, 14)
    assert.eq(service_config.max_cpu, 15)
    assert.eq(service_config.cpu_allocation, 15)
    assert.eq(service_config.min_memory, 9)
    assert.eq(service_config.max_memory, 20)
    assert.eq(service_config.memory_allocation, 20)
    assert.eq(service_config.tini_enabled, True)
    assert.eq(service_config.tolerations, [])

    # TODO
    assert.eq(service_config.user, None)

# 
def test_get_service_config_with_legacy_file(plan):
    file_name = "/key/file/0"
    file_mapping = "/value/file/0"

    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = image_name,
            files = {
                file_name: file_mapping
            }
        )
    )

    service_config = kurtosistest.get_service_config(service_name = service_name)

    # 
    # TODO Unfortunately assert is not great when comparing custom starlark types and will panic
    # 
    # That's why we need to assert the files property by property
    # 

    # First make sure the dictionary only has the specified file
    assert.eq(service_config.files.keys(), [file_name])

    # Now make sure it only contains the specified artifact
    directory = service_config.files[file_name]
    assert.eq(directory.artifact_names, [file_mapping])

def test_get_service_config_with_directory_artifacts(plan):
    directory_name = "/key/file/0"
    directory_artifact_names = ["/value/file/0", "value/file/1"]

    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = image_name,
            files = {
                directory_name: Directory(
                    artifact_names = directory_artifact_names
                )
            }
        )
    )

    service_config = kurtosistest.get_service_config(service_name = service_name)
    
    # 
    # TODO Unfortunately assert is not great when comparing custom starlark types and will panic
    # 
    # That's why we need to assert the files property by property
    # 

    # First make sure the dictionary only has the specified directory
    assert.eq(service_config.files.keys(), [directory_name])

    # Now make sure it only contains the specified artifacts
    directory = service_config.files[directory_name]
    assert.eq(directory.artifact_names, directory_artifact_names)

def test_get_service_config_with_persistent_directory_and_no_size(plan):
    directory_name = "/key/directory/0"
    persistent_key = "/persistent/key/0"

    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = image_name,
            files = {
                directory_name: Directory(
                    persistent_key = persistent_key
                )
            }
        )
    )

    service_config = kurtosistest.get_service_config(service_name = service_name)

    # 
    # TODO Unfortunately assert is not great when comparing custom starlark types and will panic
    # 
    # That's why we need to assert the files property by property
    # 

    # First make sure the dictionary only has the specified directory
    assert.eq(service_config.files.keys(), [directory_name])

    # Now check the persistent_key and the default size
    directory = service_config.files[directory_name]
    assert.eq(directory.persistent_key, persistent_key)
    assert.eq(directory.size, 1024)

def test_get_service_config_with_persistent_directory_and_size(plan):
    directory_name = "/key/directory/0"
    persistent_key = "/persistent/key/0"
    size = 3642

    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = image_name,
            files = {
                directory_name: Directory(
                    persistent_key = persistent_key,
                    size = size
                )
            }
        )
    )

    service_config = kurtosistest.get_service_config(service_name = service_name)

    # 
    # TODO Unfortunately assert is not great when comparing custom starlark types and will panic
    # 
    # That's why we need to assert the files property by property
    # 

    # First make sure the dictionary only has the specified directory
    assert.eq(service_config.files.keys(), [directory_name])

    # Now check the persistent_key and the specified size
    directory = service_config.files[directory_name]
    assert.eq(directory.persistent_key, persistent_key)
    assert.eq(directory.size, size)

def test_get_service_config_with_tolerations(plan):
    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = image_name,
            tolerations = [
                 Toleration(
                    key = "toleration-1-key",
                    value = "toleration-1-value",
                    operator = "Equal",
                    effect = "NoSchedule",
                    toleration_seconds = 64,
                ),
                Toleration(
                    # 'Toleration' expects either 'key' to be set or for 'operator' to be 'Exists'
                    operator = "Exists",
                )
            ]
        )
    )

    service_config = kurtosistest.get_service_config(service_name = service_name)

    # 
    # TODO Unfortunately assert is not great when comparing custom starlark types and will panic
    # 
    # That's why we need to assert the files property by property
    # 

    assert.eq(len(service_config.tolerations), 2)

    toleration0 = service_config.tolerations[0]
    assert.eq(toleration0.key, "toleration-1-key")
    assert.eq(toleration0.value, "toleration-1-value")
    assert.eq(toleration0.operator, "Equal")
    assert.eq(toleration0.effect, "NoSchedule")
    assert.eq(toleration0.toleration_seconds, 64)

    toleration1 = service_config.tolerations[1]
    assert.eq(toleration1.key, None)
    assert.eq(toleration1.value, None)
    assert.eq(toleration1.operator, "Exists")
    assert.eq(toleration1.effect, None)
    assert.eq(toleration1.toleration_seconds, None)