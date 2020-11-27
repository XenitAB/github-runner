all: lint fmt vet build

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -o bin/github-runner cmd/github-runner/main.go
