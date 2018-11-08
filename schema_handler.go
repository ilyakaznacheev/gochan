package gochan

//go:generate go-bindata -pkg $GOPACKAGE -o schema_types.go schema/

import (
	"bytes"
)

// GetRootSchema returnes schema from generated by go-bindata source file
func GetRootSchema() string {
	buf := bytes.Buffer{}
	for _, name := range AssetNames() {
		b := MustAsset(name)
		buf.Write(b)

		// Add a newline if the file does not end in a newline.
		if len(b) > 0 && b[len(b)-1] != '\n' {
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}