GOBINDATAEXE=$(shell go env GOPATH)/bin/go-bindata
SEDCMD=sed -i

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	SEDCMD=sed -i ""
	GOBINDATAEXE=${GOPATH}/bin/go-bindata
endif

.PHONY: all
.NOTPARALLEL: build-linux build-windows build-macos download-rclone download-rclone-windows download-rclone-macos
all: clean build-linux build-windows build-macos

.PHONY: download-rclone
download-rclone: 
	@echo "==> download rclone linux"
	@mkdir -p pkg/rclone/data/linux
	@echo "==> download rclone linux"
	@wget -L -O pkg/rclone/data/linux/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone

.PHONY: download-rclone-windows
download-rclone-windows: 
	@echo "==> download rclone windows"
	@mkdir -p pkg/rclone/data/windows
	@echo "==> download rclone windows"
	@wget -L -O pkg/rclone/data/windows/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone.exe

.PHONY: download-rclone-macos
download-rclone-macos: 
	@echo "==> download rclone macos"
	@mkdir -p pkg/rclone/data/darwin
	@echo "==> download rclone macos"
	@wget -L -O pkg/rclone/data/darwin/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone_osx

.PHONY: build-linux
build-linux:  download-rclone
	@echo "==> build sts-wire linux"
	@env GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -v -o sts-wire_linux

.PHONY: build-windows
build-windows:  download-rclone-windows
	@echo "==> build sts-wire windows"
	@env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -v -o sts-wire_windows.exe

.PHONY: build-macos
build-macos:  download-rclone-macos
	@echo "==> build sts-wire macOs"
	@env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w" -v -o sts-wire_osx

.PHONY: clean
clean:
	@echo "==> clean environment"
	@rm -f sts-wire*
	@rm -rf pkg/rclone/data/{darwin,linux,windows}
