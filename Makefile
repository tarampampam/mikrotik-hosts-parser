#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh
# LDFLAGS flags help: 'go help build'
LDFLAGS = "-s -w"
# GETFLAGS flags help: 'go help get'
GETFLAGS = -t -v -u

DOCKER_BIN = $(shell command -v docker 2> /dev/null)
DC_BIN = $(shell command -v docker-compose 2> /dev/null)
DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)" app
APP_NAME = $(notdir $(CURDIR))
GO_RUN_ARGS ?=

.PHONY : help build update gofmt test cover run shell image clean
.DEFAULT_GOAL : help
.SILENT : test shell

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-11s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

update: ## Update modules (safe)
	$(DC_BIN) run $(DC_RUN_ARGS) go get -u

build: ## Build app binary file
	$(DC_BIN) run $(DC_RUN_ARGS) go build -ldflags=$(LDFLAGS) -o '/$(APP_NAME)' .

gofmt: ## Run gofmt tool
	$(DC_BIN) run $(DC_RUN_ARGS) gofmt -s -w .

test: ## Run app tests
	$(DC_BIN) run $(DC_RUN_ARGS) go test -v -race

cover: ## Run app tests
	$(DC_BIN) run $(DC_RUN_ARGS) sh -c 'go test -v -covermode=count -coverprofile /tmp/cp.out && go tool cover -html=/tmp/cp.out -o ./coverage.html'
	-sensible-browser ./coverage.html && sleep 1 && rm -f ./coverage.html

run: ## Run app without building binary file
	$(DC_BIN) run $(DC_RUN_ARGS) go run . $(GO_RUN_ARGS)

shell: ## Start shell into container with golang
	$(DC_BIN) run $(DC_RUN_ARGS) bash

image: ## Build docker image with app
	$(DOCKER_BIN) build -f ./Dockerfile -t $(APP_NAME) .
	$(DOCKER_BIN) run $(APP_NAME) -V

clean: ## Make clean
	$(DC_BIN) down -v -t 1
	$(DOCKER_BIN) rmi $(APP_NAME) -f
