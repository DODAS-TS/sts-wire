all: build build-windows build-macos

#bindata:
#	go get -u github.com/go-bindata/go-bindata/...
#	go-bindata -o rclone_bin.go data/

#data, err := Asset("data/rclone")
#if err != nil {
#	// Asset was not found.
#}
vendors:
	go mod vendor

go-bind-data:
	go get -u github.com/go-bindata/go-bindata

bind-html: go-bind-data
	go-bindata -o pkg/core/assets.go data/html/
	sed -i "" 's/package\ main/package\ core/' pkg/core/assets.go

bind-rclone: go-bind-data
	mkdir -p data/linux
	wget -L --show-progress -O data/linux/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone
	go-bindata -o pkg/rclone/rclone_linux.go -prefix data/linux/ data/linux/
	sed -i "" 's/package\ main/package\ rclone/' pkg/rclone/rclone_linux.go

bind-rclone-windows: go-bind-data
	mkdir -p data/windows
	wget -L --show-progress -O data/windows/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone.exe
	go-bindata -o pkg/rclone/rclone_windows.go -prefix data/windows/ data/windows/
	sed -i "" 's/package\ main/package\ rclone/' pkg/rclone/rclone_windows.go

bind-rclone-macos: go-bind-data
	mkdir -p data/darwin
	wget -L --show-progress -O data/darwin/rclone https://github.com/dciangot/rclone/releases/download/v1.51.1-patch-s3/rclone_osx
	go-bindata -o pkg/rclone/rclone_darwin.go -prefix "data/darwin/" data/darwin/
	sed -i "" 's/package\ main/package\ rclone/' pkg/rclone/rclone_darwin.go

build: bind-html bind-rclone vendors
	go build -ldflags "-s -w" -o sts-wire_linux

build-windows: bind-html bind-rclone-windows vendors
	env GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -o sts-wire_windows.exe -v

build-macos: bind-html bind-rclone-macos vendors
	env GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w" -mod vendor -o sts-wire_osx -v