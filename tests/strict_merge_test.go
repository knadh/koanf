package koanf_test

import (
	"testing"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestStrictMergeFailurePreservesConfig(t *testing.T) {
	for _, conflict := range []any{123, map[string]any{"nested": true}} {
		// Map iteration order varies. Check both an early and a late conflict.
		for i := 0; i < 32; i++ {
			k := koanf.NewWithConf(koanf.Conf{Delim: ".", StrictMerge: true})
			require.NoError(t, k.Load(confmap.Provider(map[string]any{
				"config": map[string]any{
					"value": "old",
					"count": int64(1<<53 + 1),
					"list":  []int{1, 2},
				},
			}, "."), nil))
			raw, all, keys, keyMap := k.Raw(), k.All(), k.Keys(), k.KeyMap()

			err := k.Load(confmap.Provider(map[string]any{
				"config": map[string]any{
					"value": conflict,
					"count": int64(2),
					"new":   "unexpected",
				},
			}, "."), nil)
			require.ErrorContains(t, err, "incorrect types at key config.value")
			require.Equal(t, raw, k.Raw())
			require.Equal(t, all, k.All())
			require.Equal(t, keys, k.Keys())
			require.Equal(t, keyMap, k.KeyMap())
			require.False(t, k.Exists("config.new"))
			require.Nil(t, k.Get("config.new"))
			require.Equal(t, int64(1<<53+1), k.Get("config.count"))
		}
	}
}

func TestStrictMergeSuccessPreservesTypes(t *testing.T) {
	k := koanf.NewWithConf(koanf.Conf{Delim: ".", StrictMerge: true})
	require.NoError(t, k.Load(confmap.Provider(map[string]any{
		"count":  int64(1<<53 + 1),
		"list":   []int{1, 2},
		"config": map[string]any{"value": "old"},
	}, "."), nil))
	require.NoError(t, k.Load(confmap.Provider(map[string]any{
		"config": map[string]any{"value": "new", "enabled": true},
	}, "."), nil))
	require.Equal(t, int64(1<<53+1), k.Get("count"))
	require.Equal(t, []int{1, 2}, k.Get("list"))
	require.Equal(t, "new", k.Get("config.value"))
	require.True(t, k.Exists("config.enabled"))
	require.Equal(t, true, k.Get("config.enabled"))
}
