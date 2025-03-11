.PHONY: build
build:
	go build -o bin/spin-azure cmd/spin-azure/main.go

.PHONY: install
install:
	spin pluginify --install
