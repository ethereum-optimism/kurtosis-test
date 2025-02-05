# Runs go linter
# 
# With no arguments, it will lint all go workspaces
lint args="$(go list -f '{{.Dir}}/...' -m | xargs)":
    golangci-lint run {{args}}

# Runs CLI lint only
lint-cli:
    just lint ./cli/...

# Runs go build
# 
# With no arguments, it will build all go workspaces
build args="$(go list -f '{{.Dir}}/...' -m | xargs)":
    mkdir -p ./build
    go build -v -o ./build {{args}} 

# Runs CLI build only
build-cli:
    just build ./cli/...

# Runs go tests
# 
# With no arguments, it will run tests for all go workspaces
test args="$(go list -f '{{.Dir}}/...' -m | xargs)":
    go test -v {{args}}

# Runs CLI tests only
test-cli: 
    just test ./cli/...


# Cleans the build artifacts
clean:
    rm -rf build