ARCH := $(shell uname -m )

.PHONY: build
build:
	go build

.PHONY: lint
lint:
	golangci-lint run

fieldalignment:
	fieldalignment -fix ./...

package:
	fyne package -os darwin -icon assets/icon.png --appID com.changshunzhen.proxymanager

install-deps:
	go install fyne.io/fyne/v2/cmd/fyne@latest
