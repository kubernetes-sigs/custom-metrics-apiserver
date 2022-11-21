REGISTRY?=kubernetes-sigs
IMAGE?=k8s-test-metrics-adapter
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
OUT_DIR?=./_output
GOPATH:=$(shell go env GOPATH)

VERSION?=latest

GOLANGCI_VERSION:=1.50.1

.PHONY: all
all: build-test-adapter


# Generate
# --------

generated_openapis := core custommetrics externalmetrics
generated_files := $(generated_openapis:%=pkg/generated/openapi/%/zz_generated.openapi.go)

pkg/generated/openapi/core/zz_generated.openapi.go: INPUTS := k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1
pkg/generated/openapi/custommetrics/zz_generated.openapi.go: INPUTS := k8s.io/metrics/pkg/apis/custom_metrics,k8s.io/metrics/pkg/apis/custom_metrics/v1beta1,k8s.io/metrics/pkg/apis/custom_metrics/v1beta2
pkg/generated/openapi/externalmetrics/zz_generated.openapi.go: INPUTS := k8s.io/metrics/pkg/apis/external_metrics,k8s.io/metrics/pkg/apis/external_metrics/v1beta1

pkg/generated/openapi/%/zz_generated.openapi.go: go.mod go.sum
	go install -mod=readonly k8s.io/kube-openapi/cmd/openapi-gen
	$(GOPATH)/bin/openapi-gen --logtostderr \
	    -i $(INPUTS) \
	    -h ./hack/boilerplate.go.txt \
	    -p ./$(@D) \
	    -O zz_generated.openapi \
	    -o ./ \
	    -r /dev/null


# Build
# -----

.PHONY: build-test-adapter
build-test-adapter: $(generated_files)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o $(OUT_DIR)/$(ARCH)/test-adapter sigs.k8s.io/custom-metrics-apiserver/test-adapter


# Format and lint
# ---------------

HAS_GOLANGCI_VERSION:=$(shell $(GOPATH)/bin/golangci-lint version --format=short)
.PHONY: golangci
golangci:
ifneq ($(HAS_GOLANGCI_VERSION), $(GOLANGCI_VERSION))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v$(GOLANGCI_VERSION)
endif

.PHONY: verify-lint
verify-lint: golangci
	$(GOPATH)/bin/golangci-lint run --modules-download-mode=readonly || (echo 'Run "make update-lint"' && exit 1)

.PHONY: update-lint
update-lint: golangci
	$(GOPATH)/bin/golangci-lint run --fix --modules-download-mode=readonly


# Verify
# ------

.PHONY: verify
verify: verify-deps verify-lint

.PHONY: verify-deps
verify-deps:
	go mod verify
	go mod tidy
	@git diff --exit-code -- go.sum go.mod


# Test
# ----

.PHONY: test
test:
	CGO_ENABLED=0 go test ./pkg/...

.PHONY: test-adapter-container
test-adapter-container: build-test-adapter
	cp test-adapter-deploy/Dockerfile $(TEMP_DIR)
	cp $(OUT_DIR)/$(ARCH)/test-adapter $(TEMP_DIR)/adapter
	cd $(TEMP_DIR) && sed -i.bak "s|BASEIMAGE|scratch|g" Dockerfile
	sed -i.bak 's|REGISTRY|'${REGISTRY}'|g' test-adapter-deploy/testing-adapter.yaml
	docker build -t $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR) test-adapter-deploy/testing-adapter.yaml.bak

.PHONY: test-kind
test-kind:
	kind load docker-image $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION)
	kubectl apply -f test-adapter-deploy/testing-adapter.yaml
	kubectl rollout restart -n custom-metrics deployment/custom-metrics-apiserver
