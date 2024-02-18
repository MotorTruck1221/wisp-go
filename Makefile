SHELL := bash
.PHONY: all clean download run linux windows mac default compress compress-output compress-linux compress-windows compress-mac compress-all
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
	@go build -o bin/wisp-server-go -ldflags "-s -w" main.go
	@GOOS=linux GOARCH=arm go build -o bin/wisp-server-go-arm -ldflags "-s -w" main.go
	@GOOS=linux GOARCH=arm64 go build -o bin/wisp-server-go-arm64 -ldflags "-s -w" main.go
	@GOOS=linux GOARCH=386 go build -o bin/wisp-server-go-386 -ldflags "-s -w" main.go

windows: download
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o bin/wisp-server-go.exe -ldflags "-s -w" main.go

mac: download
	@echo "Building for Mac..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/wisp-server-go-mac -ldflags "-s -w" main.go
	@GOOS=darwin GOARCH=arm64 go build -o bin/wisp-server-go-mac-arm64 -ldflags "-s -w" main.go

default: clean download 
	@echo "Building for your OS..."
	@go build -o bin/wisp-server-go -ldflags "-s -w" main.go

compress:
	@echo "Compressing with no output file..."
	@upx --brute bin/wisp-server-go 

compress-output:
	@echo "Compressing with output file..."
	@upx --brute -o bin/wisp-server-go-compressed bin/wisp-server-go

compress-linux:
	@echo "Compressing for Linux..."
	@upx --brute bin/wisp-server-go --output bin/wisp-server-go-compressed
	@upx --brute bin/wisp-server-go-arm --output bin/wisp-server-go-arm-compressed
	@upx --brute bin/wisp-server-go-arm64 --output bin/wisp-server-go-arm64-compressed
	@upx --brute bin/wisp-server-go-386 --output bin/wisp-server-go-386-compressed

compress-windows:
	@echo "Compressing for Windows..."
	@upx --brute bin/wisp-server-go.exe --output bin/wisp-server-go-compressed.exe

compress-mac:
	@echo "Compressing for Mac..."
	@upx --brute bin/wisp-server-go-mac --output bin/wisp-server-go-mac-compressed
	@upx --brute bin/wisp-server-go-mac-arm64 --output bin/wisp-server-go-mac-arm64-compressed 

compress-all: compress-linux compress-windows compress-mac
	@echo "Compressing all files..."

format:
	@go fmt ./...
