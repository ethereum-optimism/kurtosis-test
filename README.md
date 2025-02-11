# kurtosis-test

Test runner for `kurtosis`

```python
my_module = import_module("/my-module")

def test_my_function(plan):
    result = my_module.my_function(plan)

    assert.true(result)
```

```bash
kurtosis-test ./my-kurtosis-package
```

## Disclaimer

> This software is in beta version, which means it is still undergoing testing and development before its official release. It may contain bugs, errors, or incomplete features that could affect its performance and functionality. By using this software, you agree to accept the risks and limitations associated with beta software. We appreciate your feedback and suggestions to help us improve this software, but we do not guarantee that we will implement them or that the software will meet your expectations. Please use this software at your own discretion and responsibility.

## Usage

`kurtosis-test` CLI currently only supports one command that runs the tests:

```bash
Usage:
  kurtosis-test <path to kurtosis project> [flags]

Flags:
  -h, --help                       help for cli
      --log-level string           Sets the level that the CLI will log at (panic|fatal|error|warning|info|debug|trace) (default "info")
      --temp-dir string            Directory for kurtosis temporary files (default ".kurtosis-test")
      --test-file-pattern string   Glob expression to use when looking for starlark test files (default "**/*_{test,spec}.star")
      --test-pattern string        Glob expression to use when looking for test functions (default "test_*")
```

## Writing starlark tests

This repository contains examples of [starlark](/test/project--passing) [tests](/test/project--failing) that are being used to test `kurtosis-test` itself.

By default, `kurtosis-test` will look for files named `*_test.star`, collecting functions named `test_*` (these also need to accept exactly one argument - the `plan` object, otherwise they will not be executed).

An example of a no-op test is:

```python
def test_not_much(plan):
    assert.true(True)
```

`kurtosis-test` comes with a built-in assertion library (under global name `assert`) and a utility module (under global name `kurtosis-test`).

### The `assert` module

The `assert` builtin module comes from [`starlarktest` package](https://github.com/google/starlark-go/blob/master/starlarktest/assert.star) and supports several useful assertions:

- `fail()`
- `fails(fn)`
- `eq(a, b)`
- `ne(a, b)`
- `lt(a, b)`
- `contains(a, b)`
- `true(a)`

### The `expect` module

Since `assert` is a reserved keyword in kurtosis, `expect` builtin is added as an alias for `assert`. The following two tests are identical:

```python
def test_with_assert(plan):
    assert.true(True)

def test_with_exoect(plan):
    expect.true(True)
```

### The `kurtosistest` module

The `kurtosistest` builtin module comes from [this repository](/cli/kurtosis/modules/kurtosistest.star). It contains functionality
that is either outside of scope of `kurtosis` or has not yet been included.

#### `kurtosistest.get_service_config(service_name)`

An extension of `plan.get_service`, `get_service_config` allows `ServiceConfig` objects to be inspected.

```python
def test_get_service_config(plan):
    plan.add_service(
        name = "my-service"
        config = ServiceConfig(
            image = "alpine:latest"
        )
    )

    service_config = kurtosistest.get_service_config(service_name = "my-service")

    assert.eq(service_config.image, "alpine:latest")
```

At the moment, only parts of the `ServiceConfig` struct are returned. The missing fields are:

- `files`
- `user`
- `tolerances`
- `ready_conditions`
- `image` field currently only returns the image name, not the build spec

#### `kurtosistest.debug(value)`

An equivalent of `print` in pure starlark, useful for debugging `kurtosistest` tests.

```python
def test_debug(plan):
    kurtosistest.debug("some value")
    kurtosistest.debug(value = "some value")
```

#### `kurtosistest.mock(target, method_name)`

Allows for spying and return value mocking of module functions:

```python
def test_mock(plan):
    # We'll create a mock object
    mock_run_sh = kurtosistest.mock(plan, "run_sh")

    # We can now mock return values
    mock_run_sh.mock_return_value("i ran sh")

    # And un-mock the return value (by passing no arguments)
    mock_run_sh.mock_return_value()

    # And inspect calls made to the mocked method
    # 
    # Every element in the calls() list is a struct containing the following fields:
    # 
    # - `args` contains a list of positional arguments
    # - `kwargs` contains a dict of named arguments
    # - `return_value` contains (mocked or unmocked) return value
    mock_run_sh.calls()
    
    # We also have access to the original method for convenience
    mock_run_sh.original
```

`kurtosistest.mock` returns a `mock` struct described above. This struct keeps track of all method calls along with their return values. It is currently not possible to restore the original method (due to the fact that every test function is run in isolation, this does not leak any mocks between tests).

## Development

### Development environment

We use [`mise`](https://mise.jdx.dev/) as a dependency manager for these tools.
Once properly installed, `mise` will provide the correct versions for each tool. `mise` does not
replace any other installations of these binaries and will only serve these binaries when you are
working inside of the `kurtosis-test` directory.

#### Install `mise`

Install `mise` by following the instructions provided on the
[Getting Started page](https://mise.jdx.dev/getting-started.html#_1-install-mise-cli).

#### Install dependencies

```sh
mise install
```