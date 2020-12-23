all: build build-windows build-macos

vendors:
	$(shell go mod vendor)

go-bindata:
	@$(shell go get -u github.com/go-bindata/go-bindata/...)

bind-html: go-bindata
	@echo "bindata html"
	@$(shell go-bindata -o pkg/core/assets.go data/html/)
	@$(shell sed -i "" "s/package\ main/package\ core/" pkg/core/assets.go)

bind-rclone: go-bindata
	@echo "download rclone linux"
	@$(shell mkdir -p data/linux)
	@$(shell wget -q -L --show-progress -O data/linux/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone)
	@echo "bindata rclone linux"
	@$(shell go-bindata -o pkg/rclone/rclone_linux.go -prefix data/linux/ data/linux/)
	@$(shell sed -i "" "s/package\ main/package\ rclone/" pkg/rclone/rclone_linux.go)

bind-rclone-windows: go-bindata
	@echo "download rclone windows"
	@$(shell mkdir -p data/windows)
	@$(shell wget -q -L --show-progress -O data/windows/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone.exe)
	@echo "bindata rclone windows"
	@$(shell go-bindata -o pkg/rclone/rclone_windows.go -prefix data/windows/ data/windows/)
	@$(shell sed -i "" "s/package\ main/package\ rclone/" pkg/rclone/rclone_windows.go)

bind-rclone-macos: go-bindata
	@echo "download rclone macos"
	@$(shell mkdir -p data/darwin)
	@$(shell wget -q -L --show-progress -O data/darwin/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone_osx)
	@echo "bindata rclone macos"
	@$(shell go-bindata -o pkg/rclone/rclone_darwin.go -prefix "data/darwin/" data/darwin/)
	@$(shell sed -i "" "s/package\ main/package\ rclone/" pkg/rclone/rclone_darwin.go)

build: bind-html bind-rclone vendors
	@echo "build sts-wire linux"
	@$(shell go build -ldflags "-s -w" -mod vendor -v -o sts-wire_linux)

build-windows: bind-html bind-rclone-windows vendors
	@echo "build sts-wire windows"
	@$(shell env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -v -o sts-wire_windows.exe)

build-macos: bind-html bind-rclone-macos vendors
	@echo "build sts-wire macOs"
	@$(shell env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -v -o sts-wire_osx)
	ls -l .