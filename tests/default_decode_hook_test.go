package koanf_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestDefaultDecodeHook(t *testing.T) {
	type config struct {
		Duration time.Duration    `koanf:"duration"`
		Count    int              `koanf:"count"`
		Value    LogFormatValue   `koanf:"value"`
		Pointer  LogFormatPointer `koanf:"pointer"`
	}

	for _, duration := range []string{"1h30m", "1d12h"} {
		t.Run(duration, func(t *testing.T) {
			k := koanf.New(".")
			require.NoError(t, k.Load(confmap.Provider(map[string]any{
				"duration": duration,
				"count":    "42",
				"value":    LogFormatValue("text_custom"),
				"pointer":  LogFormatPointer("json_custom"),
			}, "."), nil))

			var want, got config
			require.NoError(t, k.Unmarshal("", &want))
			dc := &mapstructure.DecoderConfig{
				DecodeHook:       koanf.DefaultDecodeHook(),
				WeaklyTypedInput: true,
				ErrorUnused:      true,
			}
			require.NoError(t, k.UnmarshalWithConf("", &got, koanf.UnmarshalConf{DecoderConfig: dc}))
			require.Equal(t, want, got)
			require.Equal(t, 42, got.Count)
			require.Equal(t, LogFormatValue("text_custom"), got.Value)
			require.Equal(t, LogFormatPointer("json_custom"), got.Pointer)
			if duration == "1d12h" {
				require.Equal(t, 36*time.Hour, got.Duration)
			} else {
				require.Equal(t, 90*time.Minute, got.Duration)
			}

			require.NoError(t, k.Set("unknown", true))
			require.ErrorContains(t, k.UnmarshalWithConf("", &got, koanf.UnmarshalConf{DecoderConfig: dc}), "unknown")
			// The ordinary default decoder continues to allow unused keys.
			require.NoError(t, k.Unmarshal("", &want))
		})
	}
}

func TestCustomDecoderDoesNotInheritDefaults(t *testing.T) {
	t.Run("no hooks", func(t *testing.T) {
		k := koanf.New(".")
		require.NoError(t, k.Set("duration", "1h"))
		var out struct {
			Duration time.Duration `koanf:"duration"`
		}
		require.NoError(t, k.Unmarshal("", &out))
		require.Error(t, k.UnmarshalWithConf("", &out, koanf.UnmarshalConf{
			DecoderConfig: &mapstructure.DecoderConfig{WeaklyTypedInput: true},
		}))
	})

	t.Run("no weak typing", func(t *testing.T) {
		k := koanf.New(".")
		require.NoError(t, k.Set("count", "42"))
		var out struct {
			Count int `koanf:"count"`
		}
		require.Error(t, k.UnmarshalWithConf("", &out, koanf.UnmarshalConf{
			DecoderConfig: &mapstructure.DecoderConfig{DecodeHook: koanf.DefaultDecodeHook()},
		}))
	})

	t.Run("custom hook", func(t *testing.T) {
		k := koanf.New(".")
		require.NoError(t, k.Set("duration", "custom"))
		var out struct {
			Duration time.Duration `koanf:"duration"`
		}
		hook := func(from, to reflect.Type, data any) (any, error) {
			if from.Kind() == reflect.String && to == reflect.TypeOf(time.Duration(0)) {
				return 7 * time.Hour, nil
			}
			return data, nil
		}
		require.NoError(t, k.UnmarshalWithConf("", &out, koanf.UnmarshalConf{
			DecoderConfig: &mapstructure.DecoderConfig{DecodeHook: hook},
		}))
		require.Equal(t, 7*time.Hour, out.Duration)
	})
}
