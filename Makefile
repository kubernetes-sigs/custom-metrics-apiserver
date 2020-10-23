REGISTRY?=kubernetes-sigs
IMAGE?=k8s-test-metrics-adapter
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
OUT_DIR?=./_output

OPENAPI_PATH=$(GOPATH)/src/k8s.io/kube-openapi

VERSION?=latest

ifeq  ($(PLATFORM),Linux)
SED_REPLACE :=
else
SED_REPLACE := ''
endif

.PHONY: all build-test-adapter test verify-gofmt gofmt verify test-adapter-container

all: build-test-adapter
build-test-adapter: vendor
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o $(OUT_DIR)/$(ARCH)/test-adapter github.com/kubernetes-sigs/custom-metrics-apiserver/test-adapter

vendor: tidy
	go mod vendor

tidy:
	go mod tidy

test:
	CGO_ENABLED=0 go test ./pkg/...

verify-gofmt:
	./hack/gofmt-all.sh -v

gofmt:
	./hack/gofmt-all.sh

verify: verify-gofmt test

test-adapter-container: build-test-adapter
	cp test-adapter-deploy/Dockerfile $(TEMP_DIR)
	cp $(OUT_DIR)/$(ARCH)/test-adapter $(TEMP_DIR)/adapter
	cd $(TEMP_DIR) && sed -i $(SED_REPLACE) "s|BASEIMAGE|scratch|g" Dockerfile
	sed -i $(SED_REPLACE) 's|REGISTRY|'${REGISTRY}'|g' test-adapter-deploy/testing-adapter.yaml
	docker build -t $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)

test-kind:
	kind load docker-image $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION)
	kubectl apply -f test-adapter-deploy/testing-adapter.yaml
	kubectl rollout restart -n custom-metrics deployment/custom-metrics-apiserver

test-adapter/generated/openapi/zz_generated.openapi.go:
	GO111MODULE=off go get -d k8s.io/kube-openapi || true
	rm -rf test-adapter/generated/openapi
	cd $(OPENAPI_PATH) && go run ./cmd/openapi-gen/openapi-gen.go --logtostderr -i k8s.io/metrics/pkg/apis/custom_metrics,k8s.io/metrics/pkg/apis/custom_metrics/v1beta1,k8s.io/metrics/pkg/apis/custom_metrics/v1beta2,k8s.io/metrics/pkg/apis/external_metrics,k8s.io/metrics/pkg/apis/external_metrics/v1beta1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1 -p github.com/kubernetes-sigs/custom-metrics-apiserver/test-adapter/generated/openapi -O zz_generated.openapi -r /dev/null
