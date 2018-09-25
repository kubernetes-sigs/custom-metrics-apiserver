REGISTRY?=kubernetes-incubator
IMAGE?=k8s-test-metrics-adapter
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
OUT_DIR?=./_output

VERSION?=latest

.PHONY: all build-test-adapter test verify-gofmt gofmt verify test-adapter-container

all: build-test-adapter
build-test-adapter: vendor
	CGO_ENABLED=0 GOARCH=$(ARCH) go build -o $(OUT_DIR)/$(ARCH)/test-adapter github.com/kubernetes-incubator/custom-metrics-apiserver/test-adapter

vendor: glide.lock
	glide install -v

test: vendor
	CGO_ENABLED=0 go test ./pkg/...

verify-gofmt:
	./hack/gofmt-all.sh -v

gofmt:
	./hack/gofmt-all.sh

verify: verify-gofmt test

test-adapter-container: build-test-adapter
	cp test-adapter-deploy/Dockerfile $(TEMP_DIR)
	cp $(OUT_DIR)/$(ARCH)/test-adapter $(TEMP_DIR)/adapter
	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|scratch|g" Dockerfile
	sed -i 's|REGISTRY|'${REGISTRY}'|g' test-adapter-deploy/manifests/custom-metrics-apiserver-deployment.yaml
	docker build -t $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)
