// Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// web/public/css/style.css
// +build !debug

package generated

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

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _publicCssStyleCss = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x53\xcd\x8e\xdb\x3c\x0c\x3c\xc7\x4f\x41\x60\xaf\x96\xb1\xf9\xc1\xf7\x05\x32\x90\x4b\x4f\x3d\xec\x13\x14\x3d\xd0\x16\x6d\x13\x95\x45\x43\x56\xe2\xa4\x41\xdf\xbd\x90\x92\x38\xdb\xd4\x15\x61\x1f\x46\xa2\x66\x48\x8e\x2a\x31\x17\xb8\x66\xab\x5a\xac\x78\x0d\x6f\x44\x54\x66\xab\x01\x8d\x61\xd7\xaa\x20\x83\x86\xff\xdf\x87\x73\x99\xfd\xca\xb2\x22\x60\x65\xe9\x10\x3a\x42\x73\x08\xfe\x10\xba\x1c\x1e\x58\xbc\xe7\x15\x6b\x44\xc2\x2b\x36\xe7\x9a\x85\x5c\xb3\x90\x6b\xa2\xba\x13\xf9\xc0\x35\x5a\x85\x96\x5b\xa7\xa1\x67\x63\x2c\x25\x51\x6f\x3d\xb2\x53\xb5\xb8\x80\xec\xc8\xc3\x01\x0a\x2f\xd3\xed\x77\xcd\x56\x3d\xfa\x96\x9d\xaa\x24\x04\xe9\x35\xac\xe7\x5a\xaa\xe0\xe2\xa7\x0c\x35\x78\xb4\x41\xa1\x0d\xf1\x7c\x85\xf5\x8f\xd6\xcb\xd1\x19\xf5\x68\xc9\x7f\xbb\x18\x65\xb6\xaa\xc4\x1b\xf2\x33\xbe\xdb\xc4\xf8\xe7\x6d\xba\x93\x13\xf9\xcf\xbd\x6d\xd2\x2a\x17\x49\x36\xfb\x18\x7f\x93\x6c\xb6\x31\x12\x49\x23\xbe\x07\x76\xc3\x31\x7c\xf3\x84\x46\x9c\xbd\x7c\x2f\x22\x98\xaa\xf7\x62\x3f\x73\xed\xd3\x5a\xe6\x5a\xd2\xc1\x3d\xb6\xa4\xc1\x89\xa3\x59\x84\x86\xf5\x70\x86\x51\x2c\x9b\xa7\xc0\x58\x6d\x2d\x86\xd4\xe0\xe9\xc4\x94\x9a\xdc\x88\x0b\xaa\xc1\x9e\xed\x45\xc3\x57\x57\x8b\x1b\xc5\x62\x40\xd5\xe6\x1f\xe4\xac\xe4\x1f\xe2\xb0\x96\xfc\xcb\x6d\x63\xcc\x7b\x71\x32\x0e\x58\x47\xae\x1e\xcf\x6a\x62\x13\x3a\x0d\xdb\xf7\x38\x1e\xc8\x56\x81\xce\x41\xc5\xf6\x35\x56\x26\x0d\x64\x2d\x0f\x23\x8f\x71\x6b\xea\x38\x90\x4a\xc9\x51\xee\xe4\x71\x88\xf0\xf3\x70\xc7\xc6\x90\x8b\x58\x92\x35\xf2\x4f\xd2\xb0\xde\xc6\xb9\xff\xe9\xf2\x68\x1e\x32\x1c\x24\x0d\x69\xa1\xe4\xba\xae\x9f\x03\xf1\x68\xf8\x38\x6a\xd8\xdd\x0d\x34\x78\x2a\x3a\x1e\x83\xf8\x4b\x0e\x51\x2f\x7a\xc2\xd4\x19\xb8\x66\x00\x00\x93\x78\xa3\xa2\xbc\x28\xd3\xf7\x68\xcb\x04\x77\xc4\x6d\x17\x1e\xb5\xbe\x3e\xbc\x25\x67\x6c\xee\x1e\x1b\xd0\x91\x55\xf1\x01\xb1\x6b\x81\x8b\x06\xe7\xde\xdf\x8b\x2c\x36\xd4\x97\xb3\xe7\xfd\x8d\x68\x9f\x04\xff\x0e\x00\x00\xff\xff\xb1\xe7\x33\x24\xe8\x03\x00\x00")

func publicCssStyleCssBytes() ([]byte, error) {
	return bindataRead(
		_publicCssStyleCss,
		"public/css/style.css",
	)
}

func publicCssStyleCss() (*asset, error) {
	bytes, err := publicCssStyleCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "public/css/style.css", size: 1000, mode: os.FileMode(420), modTime: time.Unix(1559359738, 0)}
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
	"public/css/style.css": publicCssStyleCss,
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
	"public": &bintree{nil, map[string]*bintree{
		"css": &bintree{nil, map[string]*bintree{
			"style.css": &bintree{publicCssStyleCss, map[string]*bintree{}},
		}},
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
