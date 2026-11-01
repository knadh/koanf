package koanf_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/knadh/koanf/providers/confmap"
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
