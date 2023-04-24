.PHONY: build test

build:
	GOARCH=arm64 GOOS=darwin go build -ldflags "-s -w" -o shortcut-darwin-arm64 ./cmd
	GOARCH=amd64 GOOS=darwin go build -ldflags "-s -w" -o shortcut-darwin-amd64 ./cmd
	GOARCH=amd64 GOOS=linux go build -ldflags "-s -w" -o shortcut-linux-amd64 ./cmd
	GOARCH=amd64 GOOS=windows go build -ldflags "-s -w" -o shortcut-windows-amd64.exe ./cmd

test:
	go test -v ./...
	go build -v ./...
	go clean
