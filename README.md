# kurtestosis

## Getting started

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

### Submodule dependencies

Install the dependencies by running the following: 
```bash
git submodule update --remote --init

# to build rvgo target
make build-rvgo

# to build rvsol target
make build-rvsol
```