def test_debug(plan):
    kurtosistest.debug(struct())
    kurtosistest.debug(value = struct())
    kurtosistest.debug(value = {
        "some": "dict"
    })

    # This will produce a warning as function pointers cannot be copied
    kurtosistest.debug(value = plan.run_sh)
