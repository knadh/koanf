package confmap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
)

func TestProvider(t *testing.T) {
	const delim = "."
	k := koanf.New(delim)
	p := confmap.Provider(map[string]interface{}{
		"foo":     nil,
		"foo.bar": "baz",
	}, delim)
	require.NoError(t, k.Load(p, nil))
	assert.Equal(t, "baz", k.String("foo.bar"))
}
