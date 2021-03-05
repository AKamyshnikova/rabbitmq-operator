SHELL:=/bin/bash
PWD := $(shell pwd)
REPOSITORY?=docker-test-local.docker.mirantis.net
REPOSITORY_PATH?=tungsten-operator
NAME?=rabbitmq-operator
VERSION?=$(shell hack/get_version.sh)
OPERATOR_IMAGE=$(REPOSITORY)/$(REPOSITORY_PATH)/$(NAME)
PUSHLATEST?=false
GOPRIVATE=gerrit.mcp.mirantis.com/*

get-version: ##Get next possible version (see hack/get_version.sh)
	@echo ${VERSION}

##@ Build

.PHONY: build
build: ## Build rabbitmq-operator executable file in local go env
	echo "Generate zzz-deepcopy objects"
	operator-sdk version
	operator-sdk generate k8s
	echo "Build RabbitMQ k8s Operator"
	operator-sdk build $(OPERATOR_IMAGE):$(VERSION) --verbose
ifeq ($(PUSHLATEST), "true")
	docker tag $(OPERATOR_IMAGE):$(VERSION) $(OPERATOR_IMAGE):latest
endif

push: ## Push rabbitmq-operator image prepared by $ make build to repository
	docker push $(OPERATOR_IMAGE):$(VERSION)
ifeq ($(PUSHLATEST), "true")
	docker push $(OPERATOR_IMAGE):latest
endif

image-path: ## Prints image path where it will be pushed
	@echo $(OPERATOR_IMAGE):$(VERSION)
image-version: ## Prints image version where it will be pushed
	@echo $(VERSION)

clean: ## Clean up the build artifacts
	@echo "Clean operator-sdk build"
	rm -rf build/_output
	@echo "Clean helm packages"
	rm -rf $(HELM_PACKAGE_DIR)/*tgz

##@ Code management

tidy: check-git-config ## Update dependencies
	go mod tidy -v

lint:
	@if golangci-lint run -v ./...; then \
	  :; \
	else \
	  code=$$?; \
	  echo "Looks like golangci-lint failed. You can try autofixes with 'make fix'."; \
	  exit $$code; \
	fi

.PHONY: fix
fix:
	golangci-lint run -v --fix ./...

check: ## Run the default dev command which is the golangci-lint then execute the $ make generate-k8s
	@echo Running the common required commands for developments purposes
	- make lint
	- make generate-k8s

check-git-config: ## Check your git config
	@if ! git --no-pager config --get-regexp 'url\..*\.insteadof' 'https://gerrit.mcp.mirantis.com/a/' 1>/dev/null; then \
		echo "go get or go tidy may fail if you don't setup Git config."; \
		echo 'To set up Git to use SSH and SSH keys auth to access Gerrit you can run:'; \
        echo '	git config --global url."ssh://$${your_login}@gerrit.mcp.mirantis.com:29418/".insteadOf "https://gerrit.mcp.mirantis.com/a/"'; \
		echo 'where $${your_login} is your login in Gerrit'; \
	fi

update-git-config: ## Update your git config
	@if ! git --no-pager config --get-regexp 'url\..*\.insteadof' 'https://gerrit.mcp.mirantis.com/a/' 1>/dev/null; then \
		git config --global url."ssh://mcp-jenkins@gerrit.mcp.mirantis.com:29418/".insteadOf "https://gerrit.mcp.mirantis.com/a/"; \
	fi

generate-k8s: ## Run the operator-sdk commands to generated code (k8s)
	@echo Updating the deep copy files with the changes in the API
	operator-sdk generate k8s

run: ## Run the development environment (in local go env) in the background using local ~/.kube/config
	export OPERATOR_NAME=rabbitmq-operator; \
	operator-sdk up local

.PHONY: help
help: ## Display this help
	@echo -e "Usage:\n  make \033[36m<target>\033[0m"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

export CGO_ENABLED=0

test: ## Run tests. Using with param example: $ make type=cover test
ifeq ($(type),cover)
	    go test ./... -cover;
else ifeq ($(type),cover-html)
		go test ./... -coverprofile coverprofile.out 1> /dev/null;
		go tool cover -html coverprofile.out;
		@rm coverprofile.out;
else ifeq ($(type),cover-func)
		go test ./... -coverprofile coverprofile.out 1> /dev/null;
		go tool cover -func coverprofile.out;
		@rm coverprofile.out;
else ifeq ($(type),cover-func-no-zero)
		go test ./... -coverprofile coverprofile.out 1> /dev/null;
		go tool cover -func coverprofile.out | grep -v -e "\t0.0%";
		@rm coverprofile.out;
else
	    go test ./...
endif
