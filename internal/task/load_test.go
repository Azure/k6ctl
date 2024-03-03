package task

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSchemaFromFile(t *testing.T) {
	t.Run("invalid file", func(t *testing.T) {
		_, err := LoadSchemaFromFile("nonexistent")
		assert.Error(t, err)
	})

	t.Run("malformed file", func(t *testing.T) {
		d := t.TempDir()
		f := filepath.Join(d, "malformed.yaml")
		assert.NoError(t, os.WriteFile(f, []byte("invalid"), 0644))

		_, err := LoadSchemaFromFile(f)
		assert.Error(t, err)
	})

	t.Run("valid file", func(t *testing.T) {
		v, err := LoadSchemaFromFile("./testdata/integration/k6ctl.yaml")
		assert.NoError(t, err)
		assert.NotNil(t, v)
	})
}
