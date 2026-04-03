.PHONY: verify test build

verify:
	gofmt -w main.go main_test.go
	go test ./...
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /tmp/llm-proxy .

test:
	go test ./...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /tmp/llm-proxy .
