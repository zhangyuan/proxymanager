.PHONY: build
build:
	go build

.PHONY: lint
lint:
	golangci-lint run

fieldalignment:
	fieldalignment -fix ./...

package:
	fyne package -os darwin -icon assets/icon.png

install-deps:
	go install fyne.io/fyne/v2/cmd/fyne@latest
