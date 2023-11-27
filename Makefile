.PHONY: clean build
.DEFAULT_GOAL := all

all: clean build

clean:
	rm -f spot
build:
	CGO_ENABLED=0 go build -ldflags '-s -w --extldflags "-static -fpic"' cmd/spot/spot.go;