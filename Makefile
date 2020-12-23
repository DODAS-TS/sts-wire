GOPATH=$(shell go env GOPATH)
BINDATA_EXE=$(GOPATH)/bin/go-bindata

all: build build-windows build-macos

vendors:
	go mod vendor

go-bindata:
	ls
	pwd
	@echo "GOPATH $(GOPATH)"
	@echo "BINDATA_EXE $(BINDATA_EXE)"
	@go get -u github.com/go-bindata/go-bindata/...
	ls $(GOPATH)
	ls $(GOBIN)

bind-html: go-bindata
	@echo "bindata html"
	$(BINDATA_EXE) -o pkg/core/assets.go data/html/
	ls pkg/core/
	@sed -i "" 's/package\ main/package\ core/' pkg/core/assets.go

bind-rclone: go-bindata
	@echo "download rclone linux"
	@mkdir -p data/linux
	@wget -q -L --show-progress -O data/linux/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone
	@echo "bindata rclone linux"
	$(BINDATA_EXE) -o pkg/rclone/rclone_linux.go -prefix data/linux/ data/linux/
	@sed -i "" 's/package\ main/package\ rclone/' pkg/rclone/rclone_linux.go

bind-rclone-windows: go-bindata
	@echo "download rclone windows"
	@mkdir -p data/windows
	@wget -q -L --show-progress -O data/windows/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone.exe
	@echo "bindata rclone windows"
	$(BINDATA_EXE) -o pkg/rclone/rclone_windows.go -prefix data/windows/ data/windows/
	@sed -i "" 's/package\ main/package\ rclone/' pkg/rclone/rclone_windows.go

bind-rclone-macos: go-bindata
	@echo "download rclone macos"
	@mkdir -p data/darwin
	@wget -q -L --show-progress -O data/darwin/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone_osx
	@echo "bindata rclone macos"
	$(BINDATA_EXE) -o pkg/rclone/rclone_darwin.go -prefix "data/darwin/" data/darwin/
	@sed -i "" 's/package\ main/package\ rclone/' pkg/rclone/rclone_darwin.go

build: bind-html bind-rclone vendors
	@echo "build sts-wire linux"
	@go build -ldflags "-s -w" -o sts-wire_linux

build-windows: bind-html bind-rclone-windows vendors
	@echo "build sts-wire windows"
	@env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -o sts-wire_windows.exe -v

build-macos: bind-html bind-rclone-macos vendors
	@echo "build sts-wire macOs"
	@env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -o sts-wire_osx -v