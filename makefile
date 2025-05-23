.DEFAULT_GOAL := run

fmt:
	go fmt ./...

build: fmt
	go build -o main-debug ./...

run: build
	./main-debug ./src/main.go
