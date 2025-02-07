# kurtestosis

Test runner for `kurtosis`

```python
my_module = import_module("/my-module")

def test_my_function(plan):
    result = my_module.my_function(plan)

    assert.true(result)
```

```bash
kurtestosis ./my-kurtosis-package
```

## Disclaimer

> This software is in beta version, which means it is still undergoing testing and development before its official release. It may contain bugs, errors, or incomplete features that could affect its performance and functionality. By using this software, you agree to accept the risks and limitations associated with beta software. We appreciate your feedback and suggestions to help us improve this software, but we do not guarantee that we will implement them or that the software will meet your expectations. Please use this software at your own discretion and responsibility.

## Usage

`kurtestosis` CLI currently only supports one command that runs the tests:

```bash
Usage:
  kurtestosis <path to kurtosis project> [flags]

Flags:
  -h, --help                       help for cli
      --log-level string           Sets the level that the CLI will log at (panic|fatal|error|warning|info|debug|trace) (default "info")
      --temp-dir string            Directory for kurtosis temporary files (default ".kurtestosis")
      --test-file-pattern string   Glob expression to use when looking for starlark test files (default "**/*_{test,spec}.star")
      --test-pattern string        Glob expression to use when looking for test functions (default "test_*")
```

## Writing starlark tests

This repository contains examples of [starlark](/test/project--passing) [tests](/test/project--failing) that are being used to test `kurtestosis` itself.

By default, `kurtestosis` will look for files named `*_test.star`, collecting functions named `test_*` (these also need to accept exactly one argument - the `plan` object, otherwise they will not be executed).

An example of a no-op test is:

```python
def test_not_much(plan):
    assert.true(True)
```

`kurtestosis` comes with a built-in assertion library (under global name `assert`) and a utility module (under global name `kurtestosis`).

### The `assert` module

The `assert` builtin module comes from [`starlarktest` package](https://github.com/google/starlark-go/blob/master/starlarktest/assert.star) and supports several useful assertions:

- `fail()`
- `fails(fn)`
- `eq(a, b)`
- `ne(a, b)`
- `lt(a, b)`
- `contains(a, b)`
- `true(a)`

### The `kurtestosis` module

The `kurtestosis` builtin module comes from [this repository](/cli/kurtosis/modules/kurtestosis.star). It contains functionality
that is either outside of scope of `kurtosis` or has not yet been included.

#### `kurtestosis.get_service_config(service_name)`

An extension of `plan.get_service`, `get_service_config` allows `ServiceConfig` objects to be inspected.

```python
def test_get_service_config(plan):
    plan.add_service(
        name = "my-service"
        config = ServiceConfig(
            image = "alpine:latest"
        )
    )

    service_config = kurtestosis.get_service_config(service_name = "my-service")

    assert.eq(service_config.image, "alpine:latest")
```

At the moment, only parts of the `ServiceConfig` struct are returned. The missing fields are:

- `files`
- `user`
- `tolerances`
- `ready_conditions`
- `image` field currently only returns the image name, not the build spec

## Development

### Development environment

We use [`mise`](https://mise.jdx.dev/) as a dependency manager for these tools.
Once properly installed, `mise` will provide the correct versions for each tool. `mise` does not
replace any other installations of these binaries and will only serve these binaries when you are
working inside of the `kurtestosis` directory.

#### Install `mise`

Install `mise` by following the instructions provided on the
[Getting Started page](https://mise.jdx.dev/getting-started.html#_1-install-mise-cli).

#### Install dependencies

```sh
mise install
```