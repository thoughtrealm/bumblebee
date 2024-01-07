BINARY_NAME=bumblebee
APP_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
APP_COMMIT := $(shell git rev-parse HEAD)
FLAGS := -X 'github.com/thoughtrealm/bumblebee/cmd.AppBuiltBy=make util' -X 'github.com/thoughtrealm/bumblebee/cmd.AppCommit=${APP_COMMIT}' -X 'github.com/thoughtrealm/bumblebee/cmd.AppDateTime=${APP_DATE}'
.DEFAULT_GOAL := build

build:
	@echo Building default platform target...
	go build -o ${BINARY_NAME} -ldflags="${FLAGS}"

install: build
	@echo
	@echo Installing Bumblebee to ${GOPATH}/bin...
	@cp ${BINARY_NAME} ${GOPATH}/bin

mac-arm64:
	@echo Building Mac ARM64 target...
	GOOS=darwin GOARCH=arm64 go build -o ${BINARY_NAME}-mac-arm64 -ldflags="${FLAGS}"

mac-amd64:
	@echo Building Mac AMD64 target...
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY_NAME}-mac-amd64 -ldflags="${FLAGS}"

win-arm64:
	@echo Building Windows ARM64 target...
	GOOS=windows GOARCH=arm64 go build -o ${BINARY_NAME}-arm64.exe -ldflags="${FLAGS}"

win-amd64:
	@echo Building Windows AMD64 target...
	GOOS=windows GOARCH=amd64 go build -o ${BINARY_NAME}-amd64.exe -ldflags="${FLAGS}"

linux-arm64:
	@echo Building Linux ARM64 target...
	GOOS=linux GOARCH=arm64 go build -o ${BINARY_NAME}-linux-arm64 -ldflags="${FLAGS}"

linux-amd64:
	@echo Building Linux AMD64 target...
	GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}-linux-amd64 -ldflags="${FLAGS}"

all: clean
	@echo
	@echo Building all targets...
	go build -o ${BINARY_NAME} -ldflags="${FLAGS}"
	@echo
	GOOS=darwin GOARCH=arm64 go build -o ${BINARY_NAME}-mac-arm64 -ldflags="${FLAGS}"
	@echo
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY_NAME}-mac-amd64 -ldflags="${FLAGS}"
	@echo
	GOOS=windows GOARCH=arm64 go build -o ${BINARY_NAME}-arm64.exe -ldflags="${FLAGS}"
	@echo
	GOOS=windows GOARCH=amd64 go build -o ${BINARY_NAME}-amd64.exe -ldflags="${FLAGS}"
	@echo
	GOOS=linux GOARCH=arm64 go build -o ${BINARY_NAME}-linux-arm64 -ldflags="${FLAGS}"
	@echo
	GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}-linux-amd64 -ldflags="${FLAGS}"

list:
	@echo Available targets
	@echo =====================
	@echo build (or just 'make') - Builds the current platform
	@echo mac-arm64
	@echo mac-amd64
	@echo win-arm64
	@echo win-amd64
	@echo linux-arm64
	@echo linux-amd64
	@echo all - builds all targets
	@echo clean - removes all targets that have been built


clean:
	@echo Cleaning GO build artifacts and removing targets...
	go clean
	-rm ${BINARY_NAME}
	-rm ${BINARY_NAME}-mac-arm64
	-rm ${BINARY_NAME}-mac-amd64
	-rm ${BINARY_NAME}-arm64.exe
	-rm ${BINARY_NAME}-amd64.exe
	-rm ${BINARY_NAME}-linux-arm64
	-rm ${BINARY_NAME}-linux-amd64
	-rm ${BINARY_NAME}-mac-arm64.tar.gz
	-rm ${BINARY_NAME}-mac-amd64.tar.gz
	-rm ${BINARY_NAME}-arm64.exe.zip
	-rm ${BINARY_NAME}-amd64.exe.zip
	-rm ${BINARY_NAME}-linux-arm64.tar.gz
	-rm ${BINARY_NAME}-linux-amd64.tar.gz

test:
	@echo Running all go tests...
	go test ./... -count=1

prep: all
	@echo
	@echo Compressing all targets...
	zip -q ${BINARY_NAME}-arm64.exe.zip ${BINARY_NAME}-arm64.exe
	zip -q ${BINARY_NAME}-amd64.exe.zip ${BINARY_NAME}-amd64.exe
	tar -czvf ${BINARY_NAME}-mac-arm64.tar.gz ${BINARY_NAME}-mac-arm64
	tar -czvf ${BINARY_NAME}-mac-amd64.tar.gz ${BINARY_NAME}-mac-amd64
	tar -czvf ${BINARY_NAME}-linux-arm64.tar.gz ${BINARY_NAME}-linux-arm64
	tar -czvf ${BINARY_NAME}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64

hashes: prep
	@echo
	@echo Generating hash file
	shasum -a 256 -b ${BINARY_NAME}-arm64.exe > builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-arm64.exe.zip >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-amd64.exe >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-amd64.exe.zip >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-mac-arm64 >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-mac-arm64.tar.gz >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-mac-amd64 >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-mac-amd64.tar.gz >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-linux-arm64 >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-linux-arm64.tar.gz >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-linux-amd64 >> builds/bee_runtimes_sha256sums.txt
	shasum -a 256 -b ${BINARY_NAME}-linux-amd64.tar.gz >> builds/bee_runtimes_sha256sums.txt
