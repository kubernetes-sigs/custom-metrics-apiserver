REGISTRY?=kubernetes-incubator
IMAGE?=k8s-custom-metric-adapter-sample
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
OUT_DIR?=./_output

VERSION?=latest

.PHONY: all build-sample test verify-gofmt gofmt verify sample-container

all: build-sample
build-sample: vendor
	CGO_ENABLED=0 GOARCH=$(ARCH) go build -o $(OUT_DIR)/$(ARCH)/sample-adapter github.com/kubernetes-incubator/custom-metrics-apiserver/test-adapter

vendor: glide.lock
	glide install -v

test: vendor
	CGO_ENABLED=0 go test ./pkg/...

verify-gofmt:
	./hack/gofmt-all.sh -v

gofmt:
	./hack/gofmt-all.sh

verify: verify-gofmt test

sample-container: build-sample
	cp sample-deploy/Dockerfile $(TEMP_DIR)
	cp $(OUT_DIR)/$(ARCH)/sample-adapter $(TEMP_DIR)/adapter
	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|scratch|g" Dockerfile
	sed -i 's|REGISTRY|'${REGISTRY}'|g' sample-deploy/manifests/custom-metrics-apiserver-deployment.yaml
	docker build -t $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)
