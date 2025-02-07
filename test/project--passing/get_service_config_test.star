def test_get_service_config(plan):
    service_name = "my-service"
    
    plan.add_service(
        name = service_name,
        config = ServiceConfig(
            image = "dependency",
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

    service_config = kurtestosis.get_service_config(service_name = service_name)

    assert.ne(service_config, None)
    assert.eq(service_config.image, "dependency")
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

    # TODO
    assert.eq(service_config.files, None)
    assert.eq(service_config.user, None)
    assert.eq(service_config.tolerations, None)