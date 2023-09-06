package koanf

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func TestLoad(t *testing.T) {
	assertFunc := func(t *testing.T, k *koanf.Koanf) {
		assert.Equal(t, "val1", k.String("key.one-example"))
		assert.Equal(t, "val2", k.String("key.two_example"))
		assert.Equal(t, 123, k.Int("key.int"))
		assert.Equal(t, 123.123, k.Float64("key.float"))
	}

	fs := &flag.FlagSet{}
	fs.String("key.one-example", "val1", "")
	fs.String("key.two_example", "val2", "")
	fs.Int("key.int", 123, "")
	fs.Float64("key.float", 123.123, "")

	k := koanf.New(".")
	assert.Nil(t, k.Load(basicflag.Provider(fs, ".", k), nil))
	assertFunc(t, k)

	// Test load with a custom key, val callback.
	k = koanf.New(".")
	p := basicflag.ProviderWithValue(fs, ".", k, func(key string, value flag.Value) (string, any) {
		if key == "key.float" {
			return "", ""
		}
		return key, value.(flag.Getter).Get()
	})
	assert.Nil(t, k.Load(p, nil), nil)
	assert.Equal(t, "val1", k.String("key.one-example"))
	assert.Equal(t, "val2", k.String("key.two_example"))
	assert.Equal(t, 123, k.Int("key.int"))
	assert.Equal(t, "", k.String("key.float"))
}

func TestLoad_Overridden(t *testing.T) {
	assertFunc := func(t *testing.T, k *koanf.Koanf) {
		// type was not set by the cli flag, but the json file provided it.
		// so it was not overridden.
		assert.Equal(t, "json", k.String("type"))
		// parent1.name was set by the cli flag, so overrides the json file value.
		assert.Equal(t, "parent1_cli_value", k.String("parent1.name"))
	}

	fs := &flag.FlagSet{}
	fs.String("type", "cli", "")
	fs.String("parent1.name", "parent1_default_value", "")

	_ = fs.Set("parent1.name", "parent1_cli_value")

	k := koanf.New(".")
	// Load JSON config.
	assert.Nil(t, k.Load(file.Provider("../mock/mock.json"), json.Parser()), nil)
	assert.Nil(t, k.Load(basicflag.Provider(fs, ".", k), nil))

	assertFunc(t, k)
}
