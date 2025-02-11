def test_debug(plan):
    kurtestosis.debug(struct())
    kurtestosis.debug(value = struct())
    kurtestosis.debug(value = {
        "some": "dict"
    })

    # This will produce a warning as function pointers cannot be copied
    kurtestosis.debug(value = plan.run_sh)
