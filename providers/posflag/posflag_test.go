package posflag_test

import (
	"strings"
	"testing"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func posflagCallback(key string, value string) (string, interface{}) {
	return strings.ReplaceAll(key, "-", "_"), value
}

func TestLoad(t *testing.T) {
	assert := func(t *testing.T, k *koanf.Koanf) {
		require.Equal(t, k.String("key.one-example"), "val1")
		require.Equal(t, k.String("key.two_example"), "val2")
		require.Equal(t, k.Strings("key.strings"), []string{"1", "2", "3"})
		require.Equal(t, k.Int("key.int"), 123)
		require.Equal(t, k.Ints("key.ints"), []int{1, 2, 3})
		require.Equal(t, k.Float64("key.float"), 123.123)
	}

	fs := &pflag.FlagSet{}
	fs.String("key.one-example", "val1", "")
	fs.String("key.two_example", "val2", "")
	fs.StringSlice("key.strings", []string{"1", "2", "3"}, "")
	fs.Int("key.int", 123, "")
	fs.IntSlice("key.ints", []int{1, 2, 3}, "")
	fs.Float64("key.float", 123.123, "")

	k := koanf.New(".")
	require.Nil(t, k.Load(posflag.Provider(fs, ".", k), nil))
	assert(t, k)

	// Test load with a custom flag callback.
	k = koanf.New(".")
	p := posflag.ProviderWithFlag(fs, ".", k, func(f *pflag.Flag) (string, interface{}) {
		return f.Name, posflag.FlagVal(fs, f)
	})
	require.Nil(t, k.Load(p, nil), nil)
	assert(t, k)

	// Test load with a custom key, val callback.
	k = koanf.New(".")
	p = posflag.ProviderWithValue(fs, ".", k, func(key, val string) (string, interface{}) {
		if key == "key.float" {
			return "", val
		}
		return key, val
	})
	require.Nil(t, k.Load(p, nil), nil)
	require.Equal(t, k.String("key.one-example"), "val1")
	require.Equal(t, k.String("key.two_example"), "val2")
	require.Equal(t, k.String("key.int"), "123")
	require.Equal(t, k.String("key.ints"), "[1,2,3]")
	require.Equal(t, k.String("key.float"), "")
}

func TestIssue90(t *testing.T) {
	exampleKeys := map[string]interface{}{
		"key.one_example": "a struct value",
		"key.two_example": "b struct value",
	}

	fs := &pflag.FlagSet{}
	fs.String("key.one-example", "a posflag value", "")
	fs.String("key.two_example", "a posflag value", "")

	k := koanf.New(".")

	err := k.Load(confmap.Provider(exampleKeys, "."), nil)
	require.Nil(t, err)

	err = k.Load(posflag.ProviderWithValue(fs, ".", k, posflagCallback), nil)
	require.Nil(t, err)

	require.Equal(t, exampleKeys, k.All())
}

func TestIssue100(t *testing.T) {
	var err error
	f := &pflag.FlagSet{}
	f.StringToString("string", map[string]string{"k": "v"}, "")
	f.StringToInt("int", map[string]int{"k": 1}, "")
	f.StringToInt64("int64", map[string]int64{"k": 2}, "")

	k := koanf.New(".")

	err = k.Load(posflag.Provider(f, ".", k), nil)
	require.Nil(t, err)

	type Maps struct {
		String map[string]string
		Int    map[string]int
		Int64  map[string]int64
	}
	maps := new(Maps)

	err = k.Unmarshal("", maps)
	require.Nil(t, err)

	require.Equal(t, map[string]string{"k": "v"}, maps.String)
	require.Equal(t, map[string]int{"k": 1}, maps.Int)
	require.Equal(t, map[string]int64{"k": 2}, maps.Int64)
}
