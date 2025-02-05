sanity_check = import_module("github.com/ethpandaops/optimism-package/src/package_io/sanity_check.star")

def test_sanity_check(plan):
    assert.fails(lambda : sanity_check.external_l1_network_params_input_parser(plan, { "value": True }), "Invalid parameter value")