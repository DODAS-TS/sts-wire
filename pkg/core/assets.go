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

var _htmlMountingpageHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x93\x51\x6f\xd3\x30\x10\xc7\x9f\xd3\x4f\x71\xeb\xc4\x1b\x69\x9a\x49\xc0\xc8\xdc\x48\x68\x1b\x62\x12\xd3\x10\x0d\xa0\x3d\xba\xce\x25\x31\x73\xec\x60\x5f\xe8\x2a\xc4\x77\x47\x8e\x9b\x36\x95\xf6\xc2\x4b\xd3\x9c\xef\xff\xbb\xbb\xbf\x2f\xec\xec\xe6\xe1\xba\x78\xfc\x72\x0b\x9f\x8a\xfb\xcf\xf9\x8c\x35\xd4\xaa\x7c\x16\xb1\x06\x79\xe9\x9f\x24\x49\x61\xbe\x2e\xd6\xf1\x8f\xbb\xaf\xb7\x2c\x09\xef\xb3\x88\xb5\x48\x1c\x1a\xa2\x2e\xc6\x5f\xbd\xfc\xbd\x9a\x5f\x1b\x4d\xa8\x29\x2e\x76\x1d\xce\x41\x84\xb7\xd5\x9c\xf0\x99\x12\x4f\xbd\x02\xd1\x70\xeb\x90\x56\xdf\x8a\x8f\xf1\xe5\x1c\x12\x8f\x71\xb4\x1b\x78\x91\x4f\x79\x3d\x8b\xa2\x8d\x29\x77\xf0\xc7\x07\x50\xd6\x0d\x65\x90\x2e\x97\xaf\xae\x66\x51\xf4\x77\x7a\x58\x4a\xd7\x29\xbe\xcb\xa0\x52\xf8\xec\x4f\xfd\x33\xde\x5a\xde\x65\xe0\x7f\x7d\xa8\xe5\xb6\x96\x3a\x83\xe5\xa8\x5e\xf8\xa1\xd0\xc6\x2d\xea\xde\x97\xaa\x8c\x21\xb4\x2f\xf3\xb8\x92\xb5\x8e\x25\x61\xeb\x32\x10\xa8\x09\xad\x0f\x6f\x65\x49\xcd\x69\x4f\x53\xea\xc0\xfa\xd9\x3b\x92\xd5\x2e\xde\x5b\x30\x95\x8f\x33\xbd\x5d\x76\x43\x95\x0d\x17\x4f\xb5\x35\xbd\x2e\x33\x38\x4f\xc5\xe5\x3b\xf1\xde\x87\x85\x51\xc6\x66\x70\x5e\x55\xd5\x58\xa5\xb9\x18\xd8\x87\x99\x60\x09\x97\x81\xe1\x4f\x7b\x05\x4a\x9e\x0e\x22\xb5\x92\x1a\xe3\x8d\x32\xe2\xc9\xa7\x75\xbc\x2c\xa5\xae\xbd\x34\xdd\x57\x57\xd2\x51\x3c\x5c\x40\x06\xda\x68\x1c\x69\x0e\x05\x49\xa3\x07\x9e\xf7\x23\x83\x74\xea\x27\xef\xc9\x4c\xbc\x78\x13\xac\x38\xf0\x47\xba\x27\x71\x4b\x52\x28\x3c\x69\xfd\x3f\xe4\x93\x0b\x7a\xa1\xfd\x13\xf3\x84\x10\x7b\x15\x4b\xc6\x9d\x62\xc9\xb8\xc5\x7e\x6f\xc6\xad\x46\x0b\x42\x71\xe7\x56\xf3\xc9\xc5\xcd\xfd\x0a\xb2\x26\x9d\x6c\x7a\x93\x1e\x10\x68\x87\x55\x0d\xb6\x0c\x99\xfb\xc1\x82\x6a\xcc\xf0\xff\x2f\xf2\x3b\x5d\x19\x96\x34\x17\x21\xe0\x5a\xae\x54\x7e\x6f\x7a\x4d\x52\xd7\xd0\x59\x23\xd0\x39\x96\x84\xb8\x97\x1f\x2b\x44\xac\xcb\xbf\x1b\xd5\xb7\x08\x5b\xa9\x14\x6c\x10\x5a\x2f\xc4\x12\xa4\x86\x0a\xb7\xe0\x50\x18\x5d\xba\xc5\x62\x01\x8f\xa6\x07\xc1\x35\x08\x65\x1c\x02\x35\xd2\x01\xf1\xcd\x19\x4b\xba\x40\x3d\x76\xc8\x92\x63\xe7\x2c\x58\x3a\xa4\x84\x16\x6e\x1e\x6e\x3e\xac\xe3\x62\x7d\x6c\x89\x25\x87\x24\x96\x04\xe7\xd8\xf0\x05\xe7\xff\x02\x00\x00\xff\xff\x25\x2b\xd9\xf0\x2e\x04\x00\x00")

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

	info := bindataFileInfo{name: "html/mountingPage.html", size: 1070, mode: os.FileMode(420), modTime: time.Unix(1609238748, 0)}
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
