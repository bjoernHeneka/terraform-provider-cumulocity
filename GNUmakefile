default: build

.PHONY: build test testacc fmt lint install

build:
	go build -v ./...

test:
	go test -v -count=1 ./internal/...

# Acceptance tests — requires TF_ACC=1 and Cumulocity credentials
# Set: CUMULOCITY_TENANT_DOMAIN, CUMULOCITY_TENANT_ID, CUMULOCITY_USERNAME, CUMULOCITY_PASSWORD
testacc:
	TF_ACC=1 go test -v -count=1 -timeout 120m ./internal/provider/...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

# Install provider locally for manual testing via ~/.terraformrc dev_overrides.
# The dev_overrides path is ~/go/bin — Terraform loads the binary from there directly.
install:
	go build -o ~/go/bin/terraform-provider-cumulocity .
