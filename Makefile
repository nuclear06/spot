.PHONY: clean build
.DEFAULT_GOAL := all
BUILD_ENV := CGO_ENABLED=0
BUILD_FLAGS := -ldflags '-s -w --extldflags "-static -fpic"'

all: clean build
release: clean release-linux-amd64 release-linux-arm64 release-windows-amd64 release-windows-arm64

clean:
	rm -f spot spot-linux-amd64 spot-linux-arm64 spot-windows-amd64.exe spot-windows-arm64.exe
build:
	${BUILD_ENV} go build ${BUILD_FLAGS} cmd/spot/spot.go;


release-linux-amd64:
	${BUILD_ENV} GOOS=linux GOARCH=amd64 go build -o spot-linux-amd64 ${BUILD_FLAGS} cmd/spot/spot.go
release-linux-arm64:
	${BUILD_ENV} GOOS=linux GOARCH=arm64 go build -o spot-linux-arm64 ${BUILD_FLAGS} cmd/spot/spot.go
release-windows-amd64:
	${BUILD_ENV} GOOS=windows GOARCH=amd64 go build -o spot-windows-amd64.exe ${BUILD_FLAGS} cmd/spot/spot.go
release-windows-arm64:
	${BUILD_ENV} GOOS=windows GOARCH=arm64 go build -o spot-windows-arm64.exe ${BUILD_FLAGS} cmd/spot/spot.go