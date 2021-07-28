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

type Example struct {
	Key ExampleKey `koanf:"key"`
}

type ExampleKey struct {
	One string `koanf:"one_example"`
	Two string `koanf:"two_example"`
}

func TestIssue90(t *testing.T) {
	exampleKeys := map[string]interface{}{
		"key.one_example": "a struct value",
		"key.two_example": "b struct value",
	}

	examplePosFlags := &pflag.FlagSet{}
	examplePosFlags.String("key.one-example", "a posflag value", "")
	examplePosFlags.String("key.two_example", "a posflag value", "")

	k := koanf.New(".")

	err := k.Load(confmap.Provider(exampleKeys, "."), nil)
	require.Nil(t, err)

	err = k.Load(posflag.ProviderWithValue(examplePosFlags, ".", k, posflagCallback), nil)
	require.Nil(t, err)

	require.Equal(t, exampleKeys, k.All())
}

type Maps struct {
	String map[string]string
	Int    map[string]int
	Int64  map[string]int64
}

func TestIssue100(t *testing.T) {
	var err error
	examplePosFlags := &pflag.FlagSet{}
	examplePosFlags.StringToString("string", map[string]string{"k": "v"}, "")
	examplePosFlags.StringToInt("int", map[string]int{"k": 1}, "")
	examplePosFlags.StringToInt64("int64", map[string]int64{"k": 2}, "")

	k := koanf.New(".")

	err = k.Load(posflag.Provider(examplePosFlags, ".", k), nil)
	require.Nil(t, err)

	maps := new(Maps)

	err = k.Unmarshal("", maps)
	require.Nil(t, err)

	require.Equal(t, map[string]string{"k": "v"}, maps.String)
	require.Equal(t, map[string]int{"k": 1}, maps.Int)
	require.Equal(t, map[string]int64{"k": 2}, maps.Int64)
}
