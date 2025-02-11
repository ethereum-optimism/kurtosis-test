mock_test_module = import_module("./mock_test_module.star")

def test_original(plan):
    mocked_method = plan.run_sh
    mock_run_sh = kurtestosis.mock(plan, "run_sh")

    assert.eq(mock_run_sh.original, mocked_method)

def test_simple(plan):
    mock_run_sh = kurtestosis.mock(plan, "run_sh")

    return_value = plan.run_sh(run = "ls")
    
    calls = mock_run_sh.calls()
    assert.eq(calls, [
        struct(args = [], kwargs = { "run": "ls" }, return_value = return_value)
    ])

def test_mock_return_value(plan):
    mock_run_sh = kurtestosis.mock(plan, "run_sh")

    mock_run_sh.mock_return_value(42)
    assert.eq(plan.run_sh(run = "pwd"), 42)

    mock_run_sh.mock_return_value(43)
    assert.eq(plan.run_sh(run = "ls"), 43)

    mock_run_sh.mock_return_value()
    return_value = plan.run_sh(run = "whoami")

    assert.ne(return_value, 42)
    assert.ne(return_value, 43)
    assert.ne(return_value, None)

    calls = mock_run_sh.calls()
    assert.eq(calls, [
        struct(args = [], kwargs = { "run": "pwd" }, return_value = 42), 
        struct(args = [], kwargs = { "run": "ls" }, return_value = 43),
        struct(args = [], kwargs = { "run": "whoami" }, return_value = return_value)
    ])

# Here we just want to make sure that mocks do not outlive the test runs
# 
# This is ensured by the fact that each test function is executed in its own thread
# by the runner
def test_mock_return_value_resets_after_test(plan):
    return_value = plan.run_sh(run = "ls")

    assert.ne(return_value, 42)
    assert.ne(return_value, 43)

def test_mock_non_builtin(plan):
    mock_module_function = kurtestosis.mock(mock_test_module, "module_function").mock_return_value(16)
    return_value = mock_test_module.module_function()
    calls = mock_module_function.calls()

    assert.eq(return_value, 16)
    assert.eq(calls, [
        struct(args = [], kwargs = {}, return_value = 16)
    ])

def test_mock_non_existing(plan):
    assert.fails(lambda: kurtestosis.mock(mock_test_module, "non_existing_function"), "doesn't have a property called non_existing_function")

def test_mock_non_function(plan):
    assert.fails(lambda: kurtestosis.mock(mock_test_module, "module_constant"), "is not a function, it's {0}".format(mock_test_module.module_constant))

def test_mock_non_module(plan):
    target = struct()
    assert.fails(lambda: kurtestosis.mock(target, "something"), "mock: only module mocks are possible at the moment. struct\\(\\) is not a module")


