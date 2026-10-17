package koanf_test

import (
	"math/big"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestDurationDecodeErrors(t *testing.T) {
	for _, input := range []string{"invalid", "1", "1w", "1d-2h", "999999999999999999999999d"} {
		k := koanf.New(".")
		require.NoError(t, k.Set("duration", input))

		var out struct{ Duration time.Duration }
		got := k.Unmarshal("", &out)
		want := k.UnmarshalWithConf("", &out, koanf.UnmarshalConf{
			DecoderConfig: &mapstructure.DecoderConfig{
				DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
			},
		})
		require.Error(t, got)
		require.Error(t, want)
		require.Equal(t, want.Error(), got.Error())
	}
}

func TestDurationDays(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  time.Duration
	}{
		{"1d", 24 * time.Hour},
		{"2d", 48 * time.Hour},
		{"1.5d", 36 * time.Hour},
		{".5d", 12 * time.Hour},
		{"1.d", 24 * time.Hour},
		{"0001.50d", 36 * time.Hour},
		{"1d12h30m", 36*time.Hour + 30*time.Minute},
		{"12h1d", 36 * time.Hour},
		{"1d2d", 72 * time.Hour},
		{"+1d", 24 * time.Hour},
		{"-1.5d", -36 * time.Hour},
		{"-1d12h", -36 * time.Hour},
		{"0.00000000000002d", time.Nanosecond},
		{"0d", 0},
		{"-0d", 0},
		{"1h30m", 90 * time.Minute},
		{"1us", time.Microsecond},
		{"1µs", time.Microsecond},
		{"0", 0},
		{"106751d23h47m16.854775807s", time.Duration(1<<63 - 1)},
		{"-106751d23h47m16.854775808s", time.Duration(-1 << 63)},
	} {
		t.Run(tt.input, func(t *testing.T) {
			k := koanf.New(".")
			require.NoError(t, k.Set("duration", tt.input))
			require.Equal(t, tt.want, k.Duration("duration"))
			if tt.want == 0 {
				require.Panics(t, func() { k.MustDuration("duration") })
			} else {
				require.Equal(t, tt.want, k.MustDuration("duration"))
			}
			var out struct{ Duration time.Duration }
			require.NoError(t, k.Unmarshal("", &out))
			require.Equal(t, tt.want, out.Duration)
		})
	}
}

func TestDurationDaysInvalid(t *testing.T) {
	for _, input := range []string{
		"", "d", ".d", "1D", "1dd", "1day", "1dms", "1d2", "1d-2h",
		"1d+2h", "1.2.3d", "1e2d", " 1d", "1d ", "1w", "NaNd", "Infd",
		"106752d", "106751d23h47m16.854775808s",
		"-106751d23h47m16.854775809s", "999999999999999999999999d",
	} {
		t.Run(input, func(t *testing.T) {
			k := koanf.New(".")
			require.NoError(t, k.Set("duration", input))
			require.Zero(t, k.Duration("duration"))
			require.Panics(t, func() { k.MustDuration("duration") })
			var out struct{ Duration time.Duration }
			require.Error(t, k.Unmarshal("", &out))
		})
	}
}

func TestDurationNumeric(t *testing.T) {
	for _, input := range []any{int64(1234), "1234", time.Duration(1234)} {
		k := koanf.New(".")
		require.NoError(t, k.Set("duration", input))
		require.Equal(t, time.Duration(1234), k.Duration("duration"))
		require.Equal(t, time.Duration(1234), k.MustDuration("duration"))
	}
}

func FuzzDurationDays(f *testing.F) {
	for _, s := range []string{"1d", ".5d", "-1d12h", "1.2.3d", "1dd", "1d-2h", "0", "1µs", "106751d23h47m16.854775807s"} {
		f.Add(s)
	}
	grammar := regexp.MustCompile(`^[+-]?(?:0|(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:ns|us|µs|μs|ms|s|m|h|d)(?:(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:ns|us|µs|μs|ms|s|m|h|d))*)$`)
	f.Fuzz(func(t *testing.T, s string) {
		k := koanf.New(".")
		require.NoError(t, k.Set("duration", s))
		var out struct{ Duration time.Duration }
		err := k.Unmarshal("", &out)
		got := out.Duration
		if err == nil && !grammar.MatchString(s) {
			t.Fatalf("accepted malformed duration %q: %v", s, got)
		}
		if !strings.Contains(s, "d") {
			want, wantErr := time.ParseDuration(s)
			if got != want || (err == nil) != (wantErr == nil) {
				t.Fatalf("changed standard parsing of %q: %v, %v; want %v, %v", s, got, err, want, wantErr)
			}
		}
	})
}

func FuzzDurationDaysDecimal(f *testing.F) {
	for _, s := range []string{"1", "1.5", ".5", "1.", "0001.50", "0.00000000000002", "999999999999999999999999.999999999999999999999", "0000000000000000000100000.000009700009999999999"} {
		f.Add(s)
	}
	decimal := regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 1000 || !decimal.MatchString(s) {
			return
		}
		want, ok := new(big.Rat).SetString(s)
		if !ok {
			t.Fatalf("invalid reference decimal %q", s)
		}
		want.Mul(want, big.NewRat(24, 1))
		precision := 0
		if dot := strings.IndexByte(s, '.'); dot >= 0 {
			precision = len(s) - dot - 1
		}

		hours := want.FloatString(precision) + "h"
		expected, expectedErr := time.ParseDuration(hours)
		k := koanf.New(".")
		require.NoError(t, k.Set("duration", s+"d"))
		var out struct{ Duration time.Duration }
		err := k.Unmarshal("", &out)
		if (err == nil) != (expectedErr == nil) || (err == nil && out.Duration != expected) {
			t.Fatalf("incorrect duration %qd: %v, %v; want %v, %v", s, out.Duration, err, expected, expectedErr)
		}
	})
}
