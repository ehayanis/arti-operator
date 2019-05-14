.PHONY: build darwin linux image release dep test

REPO= github.com/ca-gip/artifactory-operator
IMAGE ?= artifactory-operator
TAG ?= dev
DOCKER_REPO ?= silca-tools-docker-scratch-intranet.registry.saas.cagip.gca
# DOCKER_REPO ?= cagip



build:
	CGO_ENABLED=0 go build -v -o ./build/artifactory-operator -i $(GOPATH)/src/$(REPO)/cmd/main.go

darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o artifactory-operator  $(GOPATH)/src/$(REPO)/main.go

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o artifactory-operator  $(GOPATH)/src/$(REPO)/main.go

image:
	docker build --build-arg HTTP_PROXY=http://127.0.0.1:3128  --build-arg HTTPS_PROXY=http://127.0.0.1:3128 --network=host -t "$(DOCKER_REPO)/$(IMAGE):$(TAG)" .
	docker push "$(DOCKER_REPO)/$(IMAGE):$(TAG)"

release:
	docker build --build-arg HTTP_PROXY=http://127.0.0.1:3128  --build-arg HTTPS_PROXY=http://127.0.0.1:3128 --network=host -t "$(DOCKER_REPO)/$(IMAGE):$(TAG)" .
	docker push "$(DOCKER_REPO)/$(IMAGE):$(TAG)"

dep:
	glide install

test:
	go test ./... -v
