package koanf_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

type textHookValue string

func (v textHookValue) MarshalText() ([]byte, error) {
	return []byte("wire:" + string(v)), nil
}

type textHookPointer string

func (v *textHookPointer) MarshalText() ([]byte, error) {
	return []byte("wire:" + string(*v)), nil
}

type textHookDecoded string

func (v *textHookDecoded) UnmarshalText(text []byte) error {
	s := string(text)
	if !strings.HasPrefix(s, "wire:") {
		return fmt.Errorf("missing wire prefix: %q", s)
	}
	*v = textHookDecoded(strings.TrimPrefix(s, "wire:"))
	return nil
}

func TestTextUnmarshalerHookFunc(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input any
	}{
		{"plain string", "wire:value"},
		{"value marshaler", textHookValue("value")},
		{"pointer marshaler", textHookPointer("value")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			k := koanf.New(".")
			require.NoError(t, k.Set("text", tt.input))
			require.NoError(t, k.Set("duration", "1h30m"))

			type config struct {
				Text     textHookDecoded
				Duration time.Duration
			}
			var defaultOut config
			require.NoError(t, k.Unmarshal("", &defaultOut))
			require.Equal(t, textHookDecoded("value"), defaultOut.Text)
			require.Equal(t, 90*time.Minute, defaultOut.Duration)

			dc := &mapstructure.DecoderConfig{
				DecodeHook: mapstructure.ComposeDecodeHookFunc(
					mapstructure.StringToTimeDurationHookFunc(),
					koanf.TextUnmarshalerHookFunc(),
				),
				WeaklyTypedInput: true,
				ErrorUnused:      true,
			}
			var customOut config
			require.NoError(t, k.UnmarshalWithConf("", &customOut, koanf.UnmarshalConf{DecoderConfig: dc}))
			require.Equal(t, defaultOut, customOut)

			require.NoError(t, k.Set("unknown", true))
			require.ErrorContains(t, k.UnmarshalWithConf("", &customOut, koanf.UnmarshalConf{DecoderConfig: dc}), "unknown")
			require.Nil(t, dc.Result)
			require.True(t, dc.ErrorUnused)
		})
	}
}

func TestTextUnmarshalerHookFuncError(t *testing.T) {
	k := koanf.New(".")
	require.NoError(t, k.Set("text", "invalid"))
	var out struct{ Text textHookDecoded }
	err := k.UnmarshalWithConf("", &out, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{DecodeHook: koanf.TextUnmarshalerHookFunc()},
	})
	require.ErrorContains(t, err, "missing wire prefix")
}
