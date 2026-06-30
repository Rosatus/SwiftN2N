SHELL := /usr/bin/env bash

WAILS_TAGS ?= webkit2_41
WAILS_BUILD_FLAGS ?= -clean -tags "$(WAILS_TAGS)"
WAILS_DEV_FLAGS ?= -tags "$(WAILS_TAGS)"
TARGET_OS ?= $(shell go env GOOS)
TARGET_ARCH ?= $(shell go env GOARCH)
VERSION ?= dev

.PHONY: doctor test frontend-install frontend-build audit edge-linux-amd64 prepare-edge bundle-edge-linux build-current package-current build-linux release-linux dev-linux clean

doctor:
	wails doctor

test:
	go test ./...

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

audit:
	cd frontend && npm audit --audit-level=high

edge-linux-amd64:
	./scripts/download-edge-linux.sh amd64

prepare-edge:
	./scripts/prepare-edge.sh "$(TARGET_OS)" "$(TARGET_ARCH)"

bundle-edge-linux:
	mkdir -p build/bin/bin/linux/amd64
	install -m 0755 bin/linux/amd64/edge build/bin/bin/linux/amd64/edge

build-linux: edge-linux-amd64 test audit
	wails build $(WAILS_BUILD_FLAGS)
	$(MAKE) bundle-edge-linux

build-current: prepare-edge test audit
	wails build $(WAILS_BUILD_FLAGS)

package-current:
	./scripts/package-release.sh "$(TARGET_OS)" "$(TARGET_ARCH)" "$(VERSION)"

release-linux: build-linux
	rm -rf dist/SwiftN2N-linux-amd64
	mkdir -p dist/SwiftN2N-linux-amd64
	cp -a build/bin/SwiftN2N build/bin/bin dist/SwiftN2N-linux-amd64/
	cd dist && tar -czf SwiftN2N-linux-amd64.tar.gz SwiftN2N-linux-amd64

dev-linux: edge-linux-amd64
	wails dev $(WAILS_DEV_FLAGS)

clean:
	rm -rf build/bin frontend/dist dist
