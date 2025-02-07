# 
# 
# This module contains kurtestosis starlark runtime
# 
# 

# Executes a test function referenced by fn_name from module mod
# 
# The purpose of this helper function is to call the before and after hooks
# which get passed the starlark thread that is running the test
# 
# This is crucial for starlarktest go module (and its assert starlark module)
# since it requires a test reporter to be set on the thread that runs the test.
def test(plan, mod, fn_name):
    __before_test__(plan, mod, fn_name)

    fn = getattr(mod, fn_name)
    fn(plan)

    __after_test__(plan, mod, fn_name)

kurtestosis = module(
    "kurtestosis",
    test = test,
    # get_service_config is defined as a global builtin included when this file is processed
    get_service_config = get_service_config,
)