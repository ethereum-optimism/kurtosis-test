# 
# 
# This module contains kurtosistest starlark runtime
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

kurtosistest = module(
    "kurtosistest",
    test = test,
    # 
    # These are defined as global builtins when loading this module,
    # we just re-export them under the kurtosistest namespace
    # 
    get_service_config = get_service_config,
    debug = debug,
    mock = mock,
)