.PHONY: build
build:
	go build -o bin/azure cmd/azure/main.go

.PHONY: install
install:
	spin pluginify --install
