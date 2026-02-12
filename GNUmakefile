SWEEP?=all_regions
SWEEP_ACCOUNT_DIR?=./internal/services/account
SWEEP_DIR?=$(filter-out $(SWEEP_ACCOUNT_DIR), $(wildcard ./internal/services/*))
TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=scaleway
GOPATH?=$(HOME)/go

default: build

build:
	go build

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout 60m; \
	go test $(SWEEP_ACCOUNT_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout 60m

test:
	go test $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=10

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout=120m -parallel=10

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

format:
	golangci-lint fmt -v ./...

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"


test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
	@echo "Use this site to preview markdown rendering: https://registry.terraform.io/tools/doc-preview"

.PHONY: build test testacc vet format errcheck test-compile website docs

tfproviderlint:
	go tool tfproviderlint -S013=false -R014=false -AT001.ignored-filename-suffixes=_data_source_test.go ./...

tfproviderdocs:
	go tool tfproviderdocs check -provider-name scaleway -enable-contents-check

tfproviderlintx:
	go tool tfproviderlintx -S013=false -XR001=false -XS002=false ./...

gopaniccheck:
	go tool gopaniccheck ./...

format_examples:
	terraform fmt -recursive examples

docs: format_examples
	go tool tfplugindocs validate
	rm -fr ./docs
	go tool tfplugindocs generate

golangci-lint:
	golangci-lint run  ./...

typos:
	typos

lint: typos tfproviderlint tfproviderdocs tfproviderlintx golangci-lint
