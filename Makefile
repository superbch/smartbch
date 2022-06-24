VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

build_tags = cppbtree

ldflags += -X github.com/superbch/superbch/app.GitCommit=$(COMMIT) \
		  -X github.com/cosmos/cosmos-sdk/version.GitTag=$(VERSION)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
ifeq ($(OS), Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/superbchd.exe ./cmd/superbchd
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/superbchd ./cmd/superbchd
endif

build-linux: go.sum
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: all build build-linux
