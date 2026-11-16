package koanf_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntGettersNativeRange(t *testing.T) {
	for _, value := range []int64{math.MinInt32, math.MaxInt32, math.MinInt32 - 1, math.MaxInt32 + 1, 1<<32 + 1} {
		t.Run(strconv.FormatInt(value, 10), func(t *testing.T) {
			k := koanf.New(".")
			require.NoError(t, k.Load(confmap.Provider(map[string]any{
				"scalar": value,
				"text":   fmt.Sprint(value),
				"slice":  []int64{1, value},
				"mixed":  []any{1, value},
				"map":    map[string]int64{"value": value},
			}, "."), nil))

			if strconv.IntSize == 32 && (value < math.MinInt32 || value > math.MaxInt32) {
				assert.Zero(t, k.Int("scalar"))
				assert.Zero(t, k.Int("text"))
				assert.Empty(t, k.Ints("slice"))
				assert.Empty(t, k.Ints("mixed"))
				assert.Empty(t, k.IntMap("map"))
				assert.Panics(t, func() { k.MustInt("scalar") })
				assert.Panics(t, func() { k.MustInts("slice") })
				assert.Panics(t, func() { k.MustIntMap("map") })
			} else {
				assert.Equal(t, int(value), k.Int("scalar"))
				assert.Equal(t, int(value), k.Int("text"))
				assert.Equal(t, []int{1, int(value)}, k.Ints("slice"))
				assert.Equal(t, []int{1, int(value)}, k.Ints("mixed"))
				assert.Equal(t, map[string]int{"value": int(value)}, k.IntMap("map"))
			}
			assert.Equal(t, value, k.Int64("scalar"))
		})
	}
}

func TestGetNumbers(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		c.koanf = koanf.New(delim)
		require.NoError(t, c.koanf.Load(file.Provider(c.file), c.parser))

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
	}
}

func TestMustGetNumbers(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		c.koanf = koanf.New(delim)
		require.NoError(t, c.koanf.Load(file.Provider(c.file), c.parser))

		// Int.
		assert.Panics(func() { c.koanf.MustInt64("xxxx") })
		assert.Equal(int64(1234), c.koanf.MustInt64("parent1.id"))
		assert.Equal(int64(-1234), c.koanf.MustInt64("negative_int"))

		assert.Panics(func() { c.koanf.MustInt("xxxx") })
		assert.Equal(int(1234), c.koanf.MustInt("parent1.id"))
		assert.Equal(int(-1234), c.koanf.MustInt("negative_int"))

		assert.Panics(func() { c.koanf.MustInt64s("xxxx") })
		assert.Equal([]int64{1, 2, 3}, c.koanf.MustInt64s("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustInt64Map("xxxx") })
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustInt64Map("parent1.intmap"))

		assert.Panics(func() { c.koanf.MustInt64Map("parent1.boolmap") })
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustInt64Map("parent1.floatmap"))

		assert.Panics(func() { c.koanf.MustInts("xxxx") })
		assert.Equal([]int{1, 2, 3}, c.koanf.MustInts("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustIntMap("xxxx") })
		assert.Panics(func() { c.koanf.MustIntMap("parent1.boolmap") })
		assert.Equal(map[string]int{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustIntMap("parent1.intmap"))

		// Float.
		assert.Panics(func() { c.koanf.MustFloat64("xxxx") })
		assert.Equal(float64(1234), c.koanf.MustFloat64("parent1.id"))

		assert.Panics(func() { c.koanf.MustFloat64s("xxxx") })
		assert.Equal([]float64{1, 2, 3}, c.koanf.MustFloat64s("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustFloat64Map("xxxx") })
		assert.Panics(func() { c.koanf.MustFloat64Map("parent1.boolmap") })
		assert.Equal(map[string]float64{"key1": 1.1, "key2": 1.2, "key3": 1.3}, c.koanf.MustFloat64Map("parent1.floatmap"))
		assert.Equal(map[string]float64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustFloat64Map("parent1.intmap"))
	}
}

// TestConvertInt64 tests various int64 conversion/overflow scenarios.
func TestConvertInt64(t *testing.T) {
	assert := assert.New(t)
	k := koanf.New(delim)

	assert.Nil(k.Load(confmap.Provider(map[string]any{
		// Good values.
		"int64":     int64(math.MaxInt64),
		"uint64":    uint64(math.MaxInt64),
		"strmaxint": "9223372036854775807",
		"strminint": "-9223372036854775808",
		"str2p53":   "9007199254740993",
		"float":     1.9,
		"minfloat":  float64(math.MinInt64),
		"strfloat":  "3.7",

		// Overflows.
		"bigfloat":  float64(math.MaxInt64),
		"hugefloat": 1e300,
		"neghuge":   -1e300,
		"inf":       math.Inf(1),
		"neginf":    math.Inf(-1),
		"nan":       math.NaN(),
		"biguint64": uint64(math.MaxInt64) + 1,
	}, delim), nil))

	assert.Equal(int64(math.MaxInt64), k.Int64("int64"))
	assert.Equal(int64(math.MaxInt64), k.Int64("uint64"))
	assert.Equal(int64(math.MaxInt64), k.Int64("strmaxint"))
	assert.Equal(int64(math.MinInt64), k.Int64("strminint"))
	assert.Equal(int64(9007199254740993), k.Int64("str2p53"))
	assert.Equal(int64(1), k.Int64("float"))
	assert.Equal(int64(math.MinInt64), k.Int64("minfloat"))
	assert.Equal(int64(3), k.Int64("strfloat"))

	// Out of range must return 0 and not a wrapped value.
	for _, key := range []string{"bigfloat", "hugefloat", "neghuge", "inf", "neginf", "nan", "biguint64"} {
		assert.Equal(int64(0), k.Int64(key), "Int64(%q) must be 0", key)
	}
}

func TestNumberGettersNativeMaps(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
		get   func(*koanf.Koanf, string) any
	}{
		{"int", map[string]int{"a": 1, "b": 2}, func(k *koanf.Koanf, path string) any { return k.IntMap(path) }},
		{"int64", map[string]int64{"a": 1 << 54, "b": -2}, func(k *koanf.Koanf, path string) any { return k.Int64Map(path) }},
		{"float64", map[string]float64{"a": 1.5, "b": -2.5}, func(k *koanf.Koanf, path string) any { return k.Float64Map(path) }},
		{"nil_int", map[string]int(nil), func(k *koanf.Koanf, path string) any { return k.IntMap(path) }},
		{"nil_int64", map[string]int64(nil), func(k *koanf.Koanf, path string) any { return k.Int64Map(path) }},
		{"nil_float64", map[string]float64(nil), func(k *koanf.Koanf, path string) any { return k.Float64Map(path) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			k := koanf.New(delim)
			require.NoError(t, k.Load(confmap.Provider(map[string]any{"value": tt.value}, ""), nil))
			got := tt.get(k, "value")
			require.NotNil(t, got)
			if strings.HasPrefix(tt.name, "nil_") {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, tt.value, got)

			// Changing the returned map must not change the stored configuration.
			switch mp := got.(type) {
			case map[string]int:
				mp["a"] = 0
			case map[string]int64:
				mp["a"] = 0
			case map[string]float64:
				mp["a"] = 0
			}
			assert.Equal(t, tt.value, tt.get(k, "value"))
		})
	}
}
