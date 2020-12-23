GOBINDATAEXE=$(shell go env GOPATH)/bin/go-bindata

.NOTPARALLEL: build build-windows build-macos
all: build build-windows build-macos

.PHONY: vendors
vendors:
	go mod vendor

.PHONY: go-bindata-download
go-bindata-download:
	@$(shell go get -u github.com/go-bindata/go-bindata/...)

.PHONY: bind-html
bind-html: go-bindata-download
	@echo "bindata html"
	@$(shell $(GOBINDATAEXE) -o ./pkg/core/assets.go data/html/ && \
			 sed --in-place="" "s/package\\ main/package\\ core/" ./pkg/core/assets.go)

.PHONY: bind-rclone
bind-rclone: go-bindata-download
	@echo "bindata rclone linux"
	@$(shell mkdir -p data/linux && \
			 wget -q -L -O data/linux/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone && \
			 $(GOBINDATAEXE) -o ./pkg/rclone/rclone_linux.go -prefix data/linux/ ./data/linux/ && \
			 sed --in-place="" "s/package\ main/package\ rclone/" ./pkg/rclone/rclone_linux.go)

.PHONY: bind-rclone-windows
bind-rclone-windows: go-bindata-download
	@echo "bindata rclone windows"
	@$(shell mkdir -p data/windows && \
			 wget -q -L -O data/windows/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone.exe && \
			 $(GOBINDATAEXE) -o ./pkg/rclone/rclone_windows.go -prefix data/windows/ ./data/windows/ && \
			 sed --in-place="" "s/package\ main/package\ rclone/" ./pkg/rclone/rclone_windows.go)

.PHONY: bind-rclone-macos
bind-rclone-macos: go-bindata-download
	@echo "bindata rclone macos"
	@$(shell mkdir -p data/darwin && \
		wget -q -L -O data/darwin/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone_osx && \
		$(GOBINDATAEXE) -o ./pkg/rclone/rclone_darwin.go -prefix data/darwin/ ./data/darwin/ && \
		sed --in-place="" "s/package\\ main/package\\ rclone/" ./pkg/rclone/rclone_darwin.go)

.PHONY: build
build: bind-html bind-rclone vendors
	@echo "build sts-wire linux"
	@go build -ldflags "-s -w" -mod vendor -v -o sts-wire_linux

.PHONY: build-windows
build-windows: bind-html bind-rclone-windows vendors
	@echo "build sts-wire windows"
	@env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -v -o sts-wire_windows.exe

.PHONY: build-macos
build-macos: bind-html bind-rclone-macos vendors
	@echo "build sts-wire macOs"
	@env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -v -o sts-wire_osx
	ls -l .