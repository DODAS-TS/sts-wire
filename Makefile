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
.NOTPARALLEL: build-linux build-windows build-macos build-rclone-macos build-rclone-linux download-rclone download-rclone-windows download-rclone-macos
all: clean build-linux-with-rclone build-windows-with-rclone build-macos-with-rclone build-linux build-windows build-macos

.PHONY: build-rclone-windows
build-rclone-windows:
	@echo "==> bindata rclone windows"
	@powershell -Command "if (-Not (Test-Path -Path '${ROOTDIR}\pkg\rclone\data\windows')) { New-Item -ItemType Directory -Path '${ROOTDIR}\pkg\rclone\data\windows' -Force }"
	@powershell -Command "if (Test-Path -Path 'rclone') { Remove-Item -Path 'rclone' -Recurse -Force }"
	@git clone --branch rados https://github.com/DODAS-TS/rclone.git
	@echo "==> build rclone windows"
	@powershell -Command "Set-Location rclone; & '${MAKE}' build-windows; Copy-Item 'rclone\rclone$(shell go env GOEXE)' '${ROOTDIR}\pkg\rclone\data\windows\rclone$(shell go env GOEXE)'; Set-Location ${ROOTDIR}"

.PHONY: build-rclone-macos
build-rclone-macos:
	@echo "==> bindata rclone macos"
	@mkdir -p pkg/rclone/data/darwin && rm -rf rclone && git clone --branch rados https://github.com/DODAS-TS/rclone.git
	@echo "==> build rclone macos"
	@cd rclone && make build-macos && cp rclone/rclone ${ROOTDIR}/pkg/rclone/data/darwin/rclone && cd ${ROOTDIR}

.PHONY: build-rclone-linux
build-rclone-linux:
	@echo "==> bindata rclone linux"
	@mkdir -p pkg/rclone/data/linux && rm -rf rclone && git clone --branch rados https://github.com/DODAS-TS/rclone.git
	@echo "==> build rclone linux"
	@cd rclone && make build && cp rclone/rclone ${ROOTDIR}/pkg/rclone/data/linux/rclone && cd ${ROOTDIR}

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

.PHONY: build-windows-with-rclone
build-windows-with-rclone: build-rclone-windows
	@echo "==> build sts-wire windows"
	@env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.GitCommit=${GITCOMMIT}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.StsVersion=${STSVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.BuiltTime=${BUILTTIME}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.RcloneVersion=${RCLONEVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.OsArch=windows'"\
		-v -o sts-wire_windows.exe

.PHONY: build-macos-with-rclone
build-macos-with-rclone: build-rclone-macos
	@echo "==> build sts-wire macOS"
	@env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.GitCommit=${GITCOMMIT}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.StsVersion=${STSVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.BuiltTime=${BUILTTIME}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.RcloneVersion=${RCLONEVERSION}'\
		-X 'github.com/DODAS-TS/sts-wire/pkg/core.OsArch=darwin'"\
		-v -o sts-wire_osx

.PHONY: build-linux-with-rclone
build-linux-with-rclone: build-rclone-linux
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
