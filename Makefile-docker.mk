DOCKER_IMAGE := golang:1.9
INTEGRATION_DIR := nri-$(INTEGRATION)

docker-fmt:
	@echo "=== $(INTEGRATION) === [ docker-fmt ]: Running gofmt in Docker..."
	@echo "Using Docker image $(DOCKER_IMAGE)"
	@docker run -it --rm -v $(PWD):/go/src/github.com/newrelic/$(INTEGRATION_DIR) -w /go/src/github.com/newrelic/$(INTEGRATION_DIR) $(DOCKER_IMAGE) "gofmt" "-s" "-w" "."

docker-make:
	@echo "=== $(INTEGRATION) === [ docker-fmt ]: Running make in Docker..."
	@echo "Using Docker image $(DOCKER_IMAGE)"
	@docker run -it --rm -v $(PWD):/go/src/github.com/newrelic/$(INTEGRATION_DIR) -w /go/src/github.com/newrelic/$(INTEGRATION_DIR) $(DOCKER_IMAGE) "make"

.PHONY: docker-fmt docker-make