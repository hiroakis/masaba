.PHONY: all clean build

build:
	mkdir -p build
	GOARCH=386 go build -ldflags="-s -w" -o build/masaba main.go

install:
	install -m 755 build/masaba /usr/local/bin/masaba

clean:
	rm -rf ./build
