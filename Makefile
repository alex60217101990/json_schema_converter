include $(shell pwd)/deploy/makefiles/Makefile.colors
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

ROOT_REPO_PATH := `git rev-parse --show-toplevel`

##@ Docker
docker-build: ## build Dockerfile
	docker -D build --no-cache -t schema-generator -f $(shell pwd)/deploy/Dockerfile .

OVERRIDE_SCHEMA_PATH ?= "schemas/override-values.schema.json"
SCHEMA_PATH ?= "tmp/values.schema.json"
VALUES_PATH ?= ""
HELM_CHART_PATH ?= "${ROOT_REPO_PATH}"
docker-run: ## run docker container
	@printf "${COLOR_YELLOW}override schema path=${OVERRIDE_SCHEMA_PATH}${COLOR_NC}\n"
	@printf "${COLOR_YELLOW}schema path=${SCHEMA_PATH}${COLOR_NC}\n"
	@printf "${COLOR_YELLOW}values yaml path=${VALUES_PATH}${COLOR_NC}\n"
	docker run -it -v $(shell pwd)/schemas:/schemas -v "${in}":"${out}" --rm schema-generator /go/bin/app -v "${VALUES_PATH}" -s "${SCHEMA_PATH}" -o "${OVERRIDE_SCHEMA_PATH}"

##@ Run
generate: ## generate json schema with default values
	make docker-run in="${HELM_CHART_PATH}" out=/var -e OVERRIDE_SCHEMA_PATH=/schemas/override-values.schema.json -e VALUES_PATH=/var/values.yaml -e SCHEMA_PATH=/var/values.schema.json