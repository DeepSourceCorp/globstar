PACKAGE_NAME          := github.com/DeepSourceCorp/globstar
GOLANG_CROSS_VERSION  ?= v1.23

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2

CLI_BUILD_FLAGS := -X 'globstar.dev/pkg/cli.version=$$(git describe --tags 2>/dev/null || echo dev)'

.PHONY: sysroot-pack
sysroot-pack:
	@tar cf - $(SYSROOT_DIR) -P | pv -s $[$(du -sk $(SYSROOT_DIR) | awk '{print $1}') * 1024] | pbzip2 > $(SYSROOT_ARCHIVE)

.PHONY: sysroot-unpack
sysroot-unpack:
	@pv $(SYSROOT_ARCHIVE) | pbzip2 -cd | tar -xf -

.PHONY: release-dry-run
release-dry-run:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	@docker run \
		--rm \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
                release --clean --skip=publish --skip=validate

.PHONY: release
release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
                release --clean

.PHONY: test
test:
	@CGO_CFLAGS="-w" go test -coverprofile=coverage.out -covermode=atomic ./cmd/... ./pkg/...
	@go tool cover -func=coverage.out | grep total: | awk '{print "Total coverage: " $$3}'
	@rm coverage.out

.PHONY: fmt
fmt:
	@echo "Formatting Go files..."
	@gofmt -s -w .
	@echo "Done."

build:
	CGO_ENABLED=1 go build -ldflags "$(CLI_BUILD_FLAGS)" -o bin/globstar ./cmd/globstar

test-builtin-rules:
	echo "Testing built-in rules..."
	./bin/globstar test -d checkers

testall: test-builtin-rules test
