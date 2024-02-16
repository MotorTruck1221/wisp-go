SHELL := bash
.PHONY: all clean download run linux windows mac default
.DEFAULT_GOAL := default 

clean:
	@echo "Cleaning up..."
	@rm -rf bin

download:
	@echo "Downloading dependencies..."
	@go get 

run:
	@echo "Running..."
	@go run main.go start

linux: download
	@echo "Building for Linux..."
	@go build -o bin/wisp-server-go main.go
	@GOOS=linux GOARCH=arm go build -o bin/wisp-server-go-arm main.go
	@GOOS=linux GOARCH=arm64 go build -o bin/wisp-server-go-arm64 main.go
	@GOOS=linux GOARCH=386 go build -o bin/wisp-server-go-386 main.go

windows: download
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o bin/wisp-server-go.exe main.go

mac: download
	@echo "Building for Mac..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/wisp-server-go-mac main.go
	@GOOS=darwin GOARCH=arm64 go build -o bin/wisp-server-go-mac-arm64 main.go

default: clean download 
	@echo "Building for your OS..."
	@go build -o bin/wisp-server-go main.go
