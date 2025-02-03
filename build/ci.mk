.PHONY : ci/pull-builder-image
ci/pull-builder-image:
	@docker pull $(BUILDER_IMAGE)

.PHONY : ci/deps
ci/deps: ci/pull-builder-image

.PHONY : ci/debug-container
ci/debug-container: ci/deps
	@docker run --rm -it \
			--name "nri-$(INTEGRATION)-debug" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e PRERELEASE=true \
			-e GITHUB_TOKEN \
			-e REPO_FULL_NAME \
			-e TAG \
			-e GPG_MAIL \
			-e GPG_PASSPHRASE \
			-e GPG_PRIVATE_KEY_BASE64 \
			$(BUILDER_IMAGE) bash

.PHONY : ci/test
ci/test: ci/deps
	@docker run --rm -t \
			--name "nri-$(INTEGRATION)-test" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			$(BUILDER_IMAGE) make test

.PHONY : ci/snyk-test
ci/snyk-test:
	@docker run --rm -t \
			--name "nri-$(INTEGRATION)-snyk-test" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e SNYK_TOKEN \
			-e GO111MODULE=auto \
			snyk/snyk:golang snyk test --severity-threshold=high

.PHONY : ci/build
ci/build: ci/deps
ifdef TAG
	@docker run --rm -t \
			--name "nri-$(INTEGRATION)-build" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e INTEGRATION \
			-e TAG \
			$(BUILDER_IMAGE) make release/build
else
	@echo "===> $(INTEGRATION) ===  [ci/build] TAG env variable expected to be set"
	exit 1
endif

.PHONY : ci/prerelease
ci/prerelease: ci/deps
ifdef TAG
	@docker run --rm -t \
			--name "nri-$(INTEGRATION)-prerelease" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e INTEGRATION \
			-e PRERELEASE=true \
			-e GITHUB_TOKEN \
			-e REPO_FULL_NAME \
			-e TAG \
			-e GPG_MAIL \
			-e GPG_PASSPHRASE \
			-e GPG_PRIVATE_KEY_BASE64 \
			$(BUILDER_IMAGE) make release
else
	@echo "===> $(INTEGRATION) ===  [ci/prerelease] TAG env variable expected to be set"
	exit 1
endif

.PHONY : ci/fake-prerelease
ci/fake-prerelease: ci/deps
ifdef TAG
	@docker run --rm -t \
			--name "nri-$(INTEGRATION)-prerelease" \
			-v $(CURDIR):/go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-w /go/src/github.com/newrelic/nri-$(INTEGRATION) \
			-e INTEGRATION \
			-e PRERELEASE=true \
			-e NO_PUBLISH=true \
			-e NO_SIGN \
			-e GITHUB_TOKEN \
			-e REPO_FULL_NAME \
			-e TAG \
			-e GPG_MAIL \
			-e GPG_PASSPHRASE \
			-e GPG_PRIVATE_KEY_BASE64 \
			$(BUILDER_IMAGE) make release
else
	@echo "===> $(INTEGRATION) ===  [ci/fake-prerelease] TAG env variable expected to be set"
	exit 1
endif
