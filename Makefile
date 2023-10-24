SHELL       = /bin/bash
APP_NAME    = keepshare
BUILD_AT    = $(shell date +'%FT%T%z')
GO_VERSION  = $(shell go env GOVERSION)
GO_OS       = $(shell go env GOOS)
GO_ARCH     = $(shell go env GOARCH)
GIT_COMMIT  = $(shell git log --pretty=format:"%h" -1)
GIT_TAG     = $(shell git describe --abbrev=0 --tags)
VERSION     = $(shell git rev-parse --short HEAD) $(shell date +'%Y-%m-%d') $(shell go env GOVERSION)-$(shell go env GOOS)-$(shell go env GOARCH)
VERSION     = $(shell git rev-parse --short HEAD) $(shell date +'%Y-%m-%d') $(shell go env GOVERSION)-$(shell go env GOOS)-$(shell go env GOARCH)
BUILD_FLAGS = -X 'github.com/KeepShareOrg/keepshare/cmd.Version=$(GIT_TAG)' \
			  -X 'github.com/KeepShareOrg/keepshare/cmd.Commit=$(GIT_COMMIT)' \
			  -X 'github.com/KeepShareOrg/keepshare/cmd.Build=$(GO_VERSION)-$(GO_OS)-$(GO_ARCH) $(BUILD_AT)'

release: build-fe build-release
all: fmt lint test build

.PHONY: prepare
prepare:
	echo '#!/bin/sh \n\
make lint\n\
make fmt' > .git/hooks/pre-commit

.PHONY: fmt
fmt:
	gofmt -l -w .

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: build-fe
build-fe:
	if [[ -z "$$(which pnpm)" ]]; then npm install -g pnpm; fi
	cd ./web && pnpm install && npm run build

.PHONY: build
build:
	go mod tidy
	CGO_ENABLED=0 go build -ldflags="${BUILD_FLAGS}" -o ${APP_NAME} .

.PHONY: build-release
build-release:
	go mod tidy
	mkdir -p build/compress
	@for os_arch in \
		linux-amd64 linux-386 linux-arm64 linux-arm \
		windows-amd64 windows-386 \
		darwin-amd64 darwin-arm64 ; do \
		go_os=$$(echo $${os_arch}|awk -F'-' '{print $$1}') ;\
		go_arch=$$(echo $${os_arch}|awk -F'-' '{print $$2}') ;\
		name=${APP_NAME}-${GIT_TAG}-$${go_os}-$${go_arch} ;\
		echo "build $${name}" ;\
		CGO_ENABLED=0 GOOS=$${go_os} GOARCH=$${go_arch} go build -ldflags="${BUILD_FLAGS}" -o ./build/$${name} . ;\
		cd build ;\
		if [[ "$${go_os}" == "windows" ]]; then \
			exe=$${name}.exe ;\
			mv $${name} $${exe} ;\
			zip compress/$${name}.zip $${exe} >/dev/null ;\
		else \
			tar -czvf compress/$${name}.tar.gz $${name} >/dev/null ;\
		fi ;\
		cd .. ;\
	done

.PHONY: docker
docker:
	docker build .

.PHONY: gen_models
gen_models:
	go run gen_models.go --mysql="root:admin@(127.0.0.1:3306)/keepshare" --prefix=keepshare --out-path=server/query
	go run gen_models.go --mysql="root:admin@(127.0.0.1:3306)/keepshare" --prefix=pikpak --out-path=hosts/pikpak/query

.PHONY: create_table
create_table:
	make build
	./keepshare tables create --drop-empty
