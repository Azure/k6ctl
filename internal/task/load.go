package task

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

func LoadSchema(s io.Reader) (*Schema, error) {
	dec := yaml.NewDecoder(s, yaml.UseJSONUnmarshaler())

	rv := new(Schema)
	if err := dec.Decode(rv); err != nil {
		return nil, err
	}

	return rv, nil
}

func LoadSchemaFromFile(path string) (*Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadSchema(f)
}
