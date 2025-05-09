main_package_path = ./cmd/deployvia
build_dir = ./.bin
binary_name = deployvia
go_os = $(shell go env GOOS)
go_arch = $(shell go env GOARCH)

## help: Show this help message.
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## test: Run unit tests.
.PHONY: test
test:
	go test -v -cover ./...

## lint: Run linter (golangci-lint).
.PHONY: lint
lint:
	golangci-lint run ./...

## lint-fix: Run linter (golangci-lint) with auto-fix.
.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix ./...

## build: Build the binary (tries to guess the OS and architecture).
.PHONY: build
build:
	GOOS=${go_os} GOARCH=${go_arch} go build -o ${build_dir}/${binary_name} ${main_package_path}

## run: Build and then run the binary.
.PHONY: run
run: build
	LOCAL=true ${build_dir}/${binary_name}

## clean: Remove build and package directories.
.PHONY: clean
clean:
	rm -rf ${build_dir} ${package_dir}
