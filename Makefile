fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -timeout 1m ./...

build:
	go build -o bin/github-runner cmd/github-runner/main.go
