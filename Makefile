.PHONY: build test clean install

build:
	go build -o bin/promptc cmd/promptc/main.go

build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/promptc-linux-amd64 cmd/promptc/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/promptc-darwin-amd64 cmd/promptc/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/promptc-darwin-arm64 cmd/promptc/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/promptc-windows-amd64.exe cmd/promptc/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ dist/

install: build
	cp bin/promptc /usr/local/bin/

run: build
	./bin/promptc
