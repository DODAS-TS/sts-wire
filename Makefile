GOBINDATAEXE=$(shell go env GOPATH)/bin/go-bindata
SEDCMD=sed -i

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	SEDCMD=sed -i ""
	GOBINDATAEXE=${GOPATH}/bin/go-bindata
endif

.PHONY: all
.NOTPARALLEL: build-linux build-windows build-macos bind-rclone bind-rclone-windows bind-rclone-macos
all: clean build-linux build-windows build-macos

.PHONY: go-bindata-download
go-bindata-download:
	@go get -u github.com/go-bindata/go-bindata/...

.PHONY: bind-html
bind-html: go-bindata-download
	@echo "==> bindata html"
	${GOBINDATAEXE} -o pkg/core/assets.go -prefix data/ data/html/
	@echo "==> fix package"
	${SEDCMD} "s/package\ main/package\ core/g" pkg/core/assets.go

.PHONY: bind-rclone
bind-rclone: go-bindata-download
	@echo "==> bindata rclone linux"
	@mkdir -p data/linux
	@echo "==> download rclone linux"
	@wget -L -O data/linux/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone
	@echo "==> bindata rclone linux executable"
	${GOBINDATAEXE} -o pkg/rclone/rclone_linux.go -prefix data/linux/ data/linux/
	@echo "==> fix linux package"
	${SEDCMD} "s/package\ main/package\ rclone/g" pkg/rclone/rclone_linux.go

.PHONY: bind-rclone-windows
bind-rclone-windows: go-bindata-download
	@echo "==> bindata rclone windows"
	@mkdir -p data/windows
	@echo "==> download rclone windows"
	@wget -L -O data/windows/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone.exe
	@echo "==> bindata rclone windows executable"
	${GOBINDATAEXE} -o pkg/rclone/rclone_windows.go -prefix data/windows/ data/windows/
	@echo "==> fix windows package"
	${SEDCMD} "s/package\ main/package\ rclone/g" pkg/rclone/rclone_windows.go

.PHONY: bind-rclone-macos
bind-rclone-macos: go-bindata-download
	@echo "==> bindata rclone macos"
	@mkdir -p data/darwin
	@echo "==> download rclone macos"
	@wget -L -O data/darwin/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone_osx
	@echo "==> bindata rclone macos executable"
	${GOBINDATAEXE} -o pkg/rclone/rclone_darwin.go -prefix data/darwin/ data/darwin/
	@echo "==> fix macos package"
	${SEDCMD} "s/package\ main/package\ rclone/g" pkg/rclone/rclone_darwin.go

.PHONY: build-linux
build-linux: bind-html bind-rclone
	@echo "==> build sts-wire linux"
	@env GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -v -o sts-wire_linux

.PHONY: build-windows
build-windows: bind-html bind-rclone-windows
	@echo "==> build sts-wire windows"
	@env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -v -o sts-wire_windows.exe

.PHONY: build-macos
build-macos: bind-html bind-rclone-macos
	@echo "==> build sts-wire macOs"
	@env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w" -v -o sts-wire_osx

.PHONY: clean
clean:
	@echo "==> clean environment"
	@rm -f sts-wire*
	@rm -f pkg/core/assets.go
	@rm -f pkg/rclone/*.go
	@rm -rf data/{darwin,linux,windows}
