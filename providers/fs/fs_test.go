// +build go1.16

package fs_test

import (
	"fmt"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/hcl"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Case struct {
	koanf    *koanf.Koanf
	file     string
	parser   koanf.Parser
	typeName string
}

func TestFSProvider(t *testing.T) {
	var (
		assert = assert.New(t)
	)

	cases := []Case{
		{koanf: koanf.New("."), file: "mock.json", parser: json.Parser(), typeName: "json"},
		{koanf: koanf.New("."), file: "mock.yml", parser: yaml.Parser(), typeName: "yml"},
		{koanf: koanf.New("."), file: "mock.toml", parser: toml.Parser(), typeName: "toml"},
		{koanf: koanf.New("."), file: "mock.hcl", parser: hcl.Parser(true), typeName: "hcl"},
	}

	// load file system
	testFS := os.DirFS("../../mock")

	for _, c := range cases {
		// Test fs.FS before setting up kaonf
		err := fstest.TestFS(testFS, c.file)
		require.NoError(t, err, "failed asserting file existence in fs.FS")

		// koanf setup
		p := fs.Provider(testFS, c.file)
		err = c.koanf.Load(p, c.parser)
		require.NoError(t, err, fmt.Sprintf("error loading: %v", c.file))

		// Type.
		require.Equal(t, c.typeName, c.koanf.Get("type"))

		assert.Equal(nil, c.koanf.Get("xxx"))
		assert.Equal(make(map[string]interface{}), c.koanf.Get("empty"))

		// Int.
		assert.Equal(int64(0), c.koanf.Int64("xxxx"))
		assert.Equal(int64(1234), c.koanf.Int64("parent1.id"))

		assert.Equal(int(0), c.koanf.Int("xxxx"))
		assert.Equal(int(1234), c.koanf.Int("parent1.id"))

		assert.Equal([]int64{}, c.koanf.Int64s("xxxx"))
		assert.Equal([]int64{1, 2, 3}, c.koanf.Int64s("parent1.child1.grandchild1.ids"))

		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Int64Map("parent1.intmap"))
		assert.Equal(map[string]int64{}, c.koanf.Int64Map("parent1.boolmap"))
		assert.Equal(map[string]int64{}, c.koanf.Int64Map("xxxx"))
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Int64Map("parent1.floatmap"))

		assert.Equal([]int{1, 2, 3}, c.koanf.Ints("parent1.child1.grandchild1.ids"))
		assert.Equal([]int{}, c.koanf.Ints("xxxx"))

		assert.Equal(map[string]int{"key1": 1, "key2": 1, "key3": 1}, c.koanf.IntMap("parent1.intmap"))
		assert.Equal(map[string]int{}, c.koanf.IntMap("parent1.boolmap"))
		assert.Equal(map[string]int{}, c.koanf.IntMap("xxxx"))

		// Float.
		assert.Equal(float64(0), c.koanf.Float64("xxx"))
		assert.Equal(float64(1234), c.koanf.Float64("parent1.id"))

		assert.Equal([]float64{}, c.koanf.Float64s("xxxx"))
		assert.Equal([]float64{1, 2, 3}, c.koanf.Float64s("parent1.child1.grandchild1.ids"))

		assert.Equal(map[string]float64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Float64Map("parent1.intmap"))
		assert.Equal(map[string]float64{"key1": 1.1, "key2": 1.2, "key3": 1.3}, c.koanf.Float64Map("parent1.floatmap"))
		assert.Equal(map[string]float64{}, c.koanf.Float64Map("parent1.boolmap"))
		assert.Equal(map[string]float64{}, c.koanf.Float64Map("xxxx"))

		// String and bytes.
		assert.Equal([]byte{}, c.koanf.Bytes("xxxx"))
		assert.Equal([]byte("parent1"), c.koanf.Bytes("parent1.name"))

		assert.Equal("", c.koanf.String("xxxx"))
		assert.Equal("parent1", c.koanf.String("parent1.name"))

		assert.Equal([]string{}, c.koanf.Strings("xxxx"))
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.Strings("orphan"))

		assert.Equal(map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}, c.koanf.StringMap("parent1.strmap"))
		assert.Equal(map[string][]string{"key1": {"val1", "val2", "val3"}, "key2": {"val4", "val5"}}, c.koanf.StringsMap("parent1.strsmap"))
		assert.Equal(map[string]string{}, c.koanf.StringMap("xxxx"))
		assert.Equal(map[string]string{}, c.koanf.StringMap("parent1.intmap"))

		// Bools.
		assert.Equal(false, c.koanf.Bool("xxxx"))
		assert.Equal(false, c.koanf.Bool("type"))
		assert.Equal(true, c.koanf.Bool("parent1.child1.grandchild1.on"))
		assert.Equal(true, c.koanf.Bool("strbool"))

		assert.Equal([]bool{}, c.koanf.Bools("xxxx"))
		assert.Equal([]bool{true, false, true}, c.koanf.Bools("bools"))
		assert.Equal([]bool{true, false, true}, c.koanf.Bools("intbools"))
		assert.Equal([]bool{true, true, false}, c.koanf.Bools("strbools"))

		assert.Equal(map[string]bool{"ok1": true, "ok2": true, "notok3": false}, c.koanf.BoolMap("parent1.boolmap"))
		assert.Equal(map[string]bool{"key1": true, "key2": true, "key3": true}, c.koanf.BoolMap("parent1.intmap"))
		assert.Equal(map[string]bool{}, c.koanf.BoolMap("xxxx"))

		// Others.
		assert.Equal(time.Duration(1234), c.koanf.Duration("parent1.id"))
		assert.Equal(time.Duration(0), c.koanf.Duration("xxxx"))
		assert.Equal(time.Second*3, c.koanf.Duration("duration"))

		assert.Equal(time.Time{}, c.koanf.Time("xxxx", "2006-01-02"))
		assert.Equal(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), c.koanf.Time("time", "2006-01-02"))

		assert.Equal([]string{}, c.koanf.MapKeys("xxxx"), "map keys mismatch")
		assert.Equal([]string{"bools", "duration", "empty", "intbools", "negative_int", "orphan", "parent1", "parent2", "strbool", "strbools", "time", "type"},
			c.koanf.MapKeys(""), "map keys mismatch")
		assert.Equal([]string{"key1", "key2", "key3"}, c.koanf.MapKeys("parent1.strmap"), "map keys mismatch")

		// Attempt to parse int=1234 as a Unix timestamp.
		assert.Equal(time.Date(1970, 1, 1, 0, 20, 34, 0, time.UTC), c.koanf.Time("parent1.id", "").UTC())
	}
}
