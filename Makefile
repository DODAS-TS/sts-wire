ROOTDIR=$(shell git rev-parse --show-toplevel)
GOBINDATAEXE=$(shell go env GOPATH)/bin/go-bindata
SEDCMD=sed -i

GITCOMMIT = $(shell git rev-parse --short HEAD)
STSVERSION = $(shell git describe --abbrev=0 --tags)
BUILTTIME = $(shell date -u "+%Y-%m-%d %I:%M:%S%p")

RCLONEVERSION = v1.58.1-pre1
RCLONEURL = https://github.com/DODAS-TS/rclone/releases/download/

UNAME_S = $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	SEDCMD=sed -i ""
	GOBINDATAEXE=${GOPATH}/bin/go-bindata
endif

.PHONY: all
.NOTPARALLEL: build-linux build-windows build-macos build-rclone download-rclone download-rclone-windows download-rclone-macos
all: clean build-linux-with-rclone build-linux build-windows build-macos


.PHONY: build-rclone
build-rclone:
	@echo "==> bindata rclone linux"
	@mkdir -p pkg/rclone/data/linux
	@echo "==> build rclone linux"
	@rm -rf ../rclone/rclone/* && cd ../rclone/ && make build && cd ${ROOTDIR} && cp ../rclone/rclone/rclone pkg/rclone/data/linux/rclone

.PHONY: download-rclone
download-rclone:
	@echo "==> bindata rclone linux"
	@mkdir -p pkg/rclone/data/linux
	@echo "==> download rclone linux"
	@wget -L -O pkg/rclone/data/linux/rclone ${RCLONEURL}${RCLONEVERSION}/rclone_linux

.PHONY: download-rclone-windows
download-rclone-windows:
	@echo "==> bindata rclone windows"
	@mkdir -p pkg/rclone/data/windows
	@echo "==> download rclone windows"
	@wget -L -O pkg/rclone/data/windows/rclone ${RCLONEURL}${RCLONEVERSION}/rclone_windows.exe

.PHONY: download-rclone-macos
download-rclone-macos:
	@echo "==> bindata rclone macos"
	@mkdir -p pkg/rclone/data/darwin
	@echo "==> download rclone macos"
	@wget -L -O pkg/rclone/data/darwin/rclone ${RCLONEURL}${RCLONEVERSION}/rclone_osx

.PHONY: build-linux
build-linux: download-rclone
	@echo "==> build sts-wire linux"
	@env GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.GitCommit=${GITCOMMIT}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.StsVersion=${STSVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.BuiltTime=${BUILTTIME}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.RcloneVersion=${RCLONEVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.OsArch=linux'"\
		-v -o sts-wire_linux

.PHONY: build-linux-with-rclone
build-linux-with-rclone: build-rclone
	@echo "==> build sts-wire linux"
	@env GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.GitCommit=${GITCOMMIT}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.StsVersion=${STSVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.BuiltTime=${BUILTTIME}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.RcloneVersion=${RCLONEVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.OsArch=linux'"\
		-v -o sts-wire_linux

.PHONY: build-windows
build-windows: download-rclone-windows
	@echo "==> build sts-wire windows"
	@env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.GitCommit=${GITCOMMIT}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.StsVersion=${STSVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.BuiltTime=${BUILTTIME}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.RcloneVersion=${RCLONEVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.OsArch=windows'"\
		-v -o sts-wire_windows.exe

.PHONY: build-macos
build-macos: download-rclone-macos
	@echo "==> build sts-wire macOS"
	@env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w \
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.GitCommit=${GITCOMMIT}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.StsVersion=${STSVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.BuiltTime=${BUILTTIME}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.RcloneVersion=${RCLONEVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.OsArch=darwin'"\
		-v -o sts-wire_osx

.PHONY: clean
clean:
	@echo "==> clean environment"
	@rm -f sts-wire*
	@rm -rf pkg/rclone/data
