def test_debug(plan):
    kurtosistestdebug(struct())
    kurtosistestdebug(value = struct())
    kurtosistestdebug(value = {
        "some": "dict"
    })

    # This will produce a warning as function pointers cannot be copied
    kurtosistestdebug(value = plan.run_sh)
