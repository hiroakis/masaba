.PHONY: all clean deps test build install

all: clean deps test build

deps:
	go get -u github.com/stretchr/testify/assert

test:
	go test -v

build: deps test
	mkdir -p build
	GOARCH=386 go build -ldflags="-s -w" -o build/masaba main.go

install:
	install -m 755 build/masaba /usr/local/bin/masaba

clean:
	rm -rf ./build
