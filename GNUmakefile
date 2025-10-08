default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test $(GO_TEST_FLAGS) -parallel=10 $(PKGS)

provider-testacc:
	TF_ACC=1 go test $(GO_TEST_FLAGS) $(PROVIDER_PKGS)

provider-testacc-tofu:
	@set -e; \
	BIN="$${TOFU_BIN:-$$(command -v tofu || true)}"; \
	if [ -z "$$BIN" ]; then echo "ERROR: tofu not found on PATH (or TOFU_BIN unset)"; exit 1; fi; \
	echo "Using tofu at $$BIN"; \
	TF_ACC=1 TF_ACC_PROVIDER_HOST="registry.opentofu.org" TF_ACC_TERRAFORM_PATH="$$BIN" go test $(GO_TEST_FLAGS) $(PROVIDER_PKGS) $(ARGS)


.PHONY: fmt lint test provider-testacc provider-testacc-tofu build install generate

GO_TEST_FLAGS = -v -cover -timeout 120m
PROVIDER_PKGS = ./internal/provider
PKGS ?= ./...
