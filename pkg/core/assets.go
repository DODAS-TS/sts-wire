// Code generated for package core by go-bindata DO NOT EDIT. (@generated)
// sources:
// data/html/mountingPage.html
package core

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _htmlMountingpageHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x53\xc1\x6e\xdb\x3a\x10\x3c\xcb\x5f\xb1\x71\xf0\x6e\x4f\x96\x1d\xa0\x6d\xaa\xd0\x02\x8a\x24\x45\x03\x34\x48\x51\xbb\x2d\x72\xa4\xa9\x95\xc4\x86\x22\x55\x72\x55\xc7\x28\xf2\xef\xc5\x4a\x96\x2d\x03\x39\xf4\x62\x59\xcb\x9d\xd9\x9d\xe1\x48\x9c\xdd\x3c\x5c\xaf\x1f\xbf\xdc\xc2\xa7\xf5\xfd\xe7\x6c\x22\x2a\xaa\x4d\x36\x89\x44\x85\x32\xe7\x27\x69\x32\x98\xad\xd6\xab\xf8\xc7\xdd\xd7\x5b\x91\xf4\xef\x93\x48\xd4\x48\x12\x2a\xa2\x26\xc6\x5f\xad\xfe\xbd\x9c\x5e\x3b\x4b\x68\x29\x5e\xef\x1a\x9c\x82\xea\xdf\x96\x53\xc2\x67\x4a\x98\xf5\x0a\x54\x25\x7d\x40\x5a\x7e\x5b\x7f\x8c\x2f\xa7\x90\x30\x4d\xa0\x5d\xc7\x17\x71\xcb\xff\x93\x28\xda\xb8\x7c\x07\x7f\xb8\x80\xba\xac\x28\x85\xc5\x7c\xfe\xdf\xd5\x24\x8a\x5e\xc6\x87\xb9\x0e\x8d\x91\xbb\x14\x0a\x83\xcf\x7c\xca\xcf\x78\xeb\x65\x93\x02\xff\x72\xa9\x96\xbe\xd4\x36\x85\xf9\x80\x9e\xb1\x28\xf4\x71\x8d\xb6\xe5\x51\x85\x73\x84\xfe\x75\x3e\x69\x74\x69\x63\x4d\x58\x87\x14\x14\x5a\x42\xcf\xe5\xad\xce\xa9\x3a\xdd\x69\xcc\xda\x71\xfd\x6c\x03\xe9\x62\x17\xef\x2d\x18\xc3\x07\x4d\x6f\xe7\x4d\x37\x65\x23\xd5\x53\xe9\x5d\x6b\xf3\x14\xce\x17\xea\xf2\x9d\x7a\xcf\x65\xe5\x8c\xf3\x29\x9c\x17\x45\x31\x4c\xa9\x2e\x3a\xee\x83\x26\x98\xc3\x65\xcf\xc1\xa7\xad\x01\xa3\x4f\x85\x68\x6b\xb4\xc5\x78\x63\x9c\x7a\xe2\xb6\x46\xe6\xb9\xb6\x25\x43\x17\xfb\xe9\x46\x07\x8a\xbb\x0b\x48\xc1\x3a\x8b\x03\x5b\x40\x45\xda\xd9\x8e\x8f\xfd\x48\x61\x31\xd2\xfe\xa6\x97\x7e\xe0\x1b\xd8\x18\x29\x3d\x69\x65\xf0\x64\x55\xd9\x92\xfb\x07\x38\xe7\x24\xee\x5c\x1f\x1b\xf6\x72\x7a\x4d\xaf\x88\x38\xb1\x30\xcf\xf3\x3d\x4a\x24\x43\xb2\x44\x32\x64\x99\xd3\x33\x64\x1b\x3d\x28\x23\x43\x58\x4e\x47\xd7\x37\xe5\x20\x8a\x6a\x31\xca\x7b\xb5\x38\x50\xa0\xef\x02\xdb\x9b\xd3\x75\xee\xe5\xf6\xa8\xa1\x83\xff\x5f\x64\x77\xb6\x70\x22\xa9\x2e\xfa\x42\xa8\xa5\x31\xd9\xbd\x6b\x2d\x69\x5b\x42\xe3\x9d\xc2\x10\x44\xd2\xd7\x19\x7e\x9c\x10\x89\x26\xfb\xee\x4c\x5b\x23\x6c\xb5\x31\xb0\x41\xa8\x19\x88\x39\x68\x0b\x05\x6e\x21\xa0\x72\x36\x0f\xb3\xd9\x4c\x6c\x7c\xf6\xe8\x5a\x50\xd2\x82\x32\x2e\x20\x50\xa5\x03\x90\xdc\x9c\x89\xa4\xe9\x89\x8f\x4b\x8a\xe4\xb8\xbc\xe8\x5d\xed\x5a\xfa\x2d\x6e\x1e\x6e\x3e\xac\xe2\xf5\xea\xb8\x95\x48\x0e\x4d\x22\xe9\xcd\x13\xdd\xa7\x9c\xfd\x0d\x00\x00\xff\xff\x36\xd8\xb9\x43\x37\x04\x00\x00")

func htmlMountingpageHtmlBytes() ([]byte, error) {
	return bindataRead(
		_htmlMountingpageHtml,
		"html/mountingPage.html",
	)
}

func htmlMountingpageHtml() (*asset, error) {
	bytes, err := htmlMountingpageHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "html/mountingPage.html", size: 1079, mode: os.FileMode(420), modTime: time.Unix(1609243665, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"html/mountingPage.html": htmlMountingpageHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"html": &bintree{nil, map[string]*bintree{
		"mountingPage.html": &bintree{htmlMountingpageHtml, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
