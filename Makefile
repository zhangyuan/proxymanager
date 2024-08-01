.PHONY: build
build:
	go build -o bin/proxy-manager

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/proxy-manager_linux-amd64 -ldflags="-w -s -buildid=" -trimpath

.PHONY: build-maco
build-macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/proxy-manager_darwin-amd64 -ldflags="-w -s -buildid=" -trimpath
	GOOS=darwin GOARCH=arm64 go build -o bin/proxy-manager_darwin-arm64 -ldflags="-w -s -buildid=" -trimpath

.PHONY: build-all
build-all: build-macos build-linux

.PHONY: serve
serve:
	go run main.go serve

.PHONY: lint
lint:
	golangci-lint run

fieldalignment:
	fieldalignment -fix ./...

package:
	fyne package -os darwin -icon assets/icon.png
