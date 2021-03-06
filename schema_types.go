// Code generated by go-bindata.
// sources:
// schema/schema.graphql
// DO NOT EDIT!

package gochan

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

var _schemaSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x52\x4d\x4f\xc3\x30\x0c\xbd\xe7\x57\x78\xda\x65\x93\xf8\x05\xb9\x81\x76\xa0\x12\x48\xe3\xeb\x84\x76\x08\x34\x6a\x23\xad\x4b\x69\x5d\xc1\x84\xf6\xdf\x91\x1d\x27\x69\x60\xec\xd4\xda\x79\xef\xd9\xef\x25\xe3\x7b\x6b\x3b\x03\xdf\x0a\x00\xe0\x63\xb2\xc3\x51\xc3\x03\x7d\xb8\xd1\x4d\x68\xd0\xf9\x83\x86\x7b\xf9\x53\x27\xa5\xf0\xd8\xdb\x00\x12\xde\x12\x5a\xdf\x59\xe8\x4d\x63\xe1\xd3\x61\x0b\x6f\xde\x0c\x35\xec\xdd\x88\x7c\xde\x58\xbc\xf5\x9d\x5d\xad\x35\xd0\x57\x38\x01\x94\x49\xd8\x0e\xd6\x94\xac\x1b\x82\xac\x5c\xad\xe1\x09\x07\x77\x68\x16\x6b\x0d\xdc\x13\x09\xa1\x64\x8d\xde\x8f\x58\x28\x3c\x33\x82\x25\xaa\x0d\xd1\x43\x43\xf8\x04\x8f\xc8\xad\x1f\x71\x86\xdb\xc6\xa3\x25\x98\x09\x5b\x3f\x5c\x9a\x72\xcd\x88\x5f\x8b\x86\x66\x0a\x2c\x26\x28\x99\x99\xba\xe6\x89\xc1\x42\xb5\xe1\xb9\x57\x2c\x1d\x86\x57\x87\x7e\xc2\x62\x13\x53\xd7\x62\x87\xa3\x4b\x9c\x20\x11\xad\x25\x9e\x38\x8d\xf3\x29\x79\x99\xcd\xf4\x51\xc3\x2b\x67\xb9\x4b\x10\x2e\x05\x93\xbd\x70\x89\x0e\xf7\xb6\xec\xb0\x3c\xa9\x84\x41\x59\x26\xd4\x33\x9d\x6a\xf3\x8f\x46\xcb\x7b\x27\x83\x64\x9e\x04\xa9\xb1\x0b\x96\x39\xc3\x3f\x59\x12\xe0\x8c\xbe\xfd\xc2\x42\xde\x75\x8d\x86\xaa\x33\x8d\xbd\x24\x16\xca\xf3\xb6\xcb\x8d\x22\x83\x25\x85\xf0\xf2\x78\x97\x18\x27\xa5\x1c\xa5\x9f\xef\x4f\x40\xf3\xcd\x16\xf2\xaa\xf2\x72\xe1\xc6\x32\x7b\x76\x8f\x91\x3f\x4f\x6e\x91\x36\x9b\x3f\x14\xa2\xab\x9f\x00\x00\x00\xff\xff\xad\x1d\x50\xc8\xce\x03\x00\x00")

func schemaSchemaGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_schemaSchemaGraphql,
		"schema/schema.graphql",
	)
}

func schemaSchemaGraphql() (*asset, error) {
	bytes, err := schemaSchemaGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/schema.graphql", size: 974, mode: os.FileMode(420), modTime: time.Unix(1541702767, 0)}
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
	"schema/schema.graphql": schemaSchemaGraphql,
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
	"schema": &bintree{nil, map[string]*bintree{
		"schema.graphql": &bintree{schemaSchemaGraphql, map[string]*bintree{}},
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

