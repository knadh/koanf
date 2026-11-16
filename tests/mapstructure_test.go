package koanf_test

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parentStruct struct {
	Name   string      `koanf:"name"`
	ID     int         `koanf:"id"`
	Child1 childStruct `koanf:"child1"`
}
type childStruct struct {
	Name        string            `koanf:"name"`
	Type        string            `koanf:"type"`
	Empty       map[string]string `koanf:"empty"`
	Grandchild1 grandchildStruct  `koanf:"grandchild1"`
}
type grandchildStruct struct {
	Ids []int `koanf:"ids"`
	On  bool  `koanf:"on"`
}
type testStruct struct {
	Type    string            `koanf:"type"`
	Empty   map[string]string `koanf:"empty"`
	Parent1 parentStruct      `koanf:"parent1"`
}

type testStructFlat struct {
	Type                        string            `koanf:"type"`
	Empty                       map[string]string `koanf:"empty"`
	Parent1Name                 string            `koanf:"parent1.name"`
	Parent1ID                   int               `koanf:"parent1.id"`
	Parent1Child1Name           string            `koanf:"parent1.child1.name"`
	Parent1Child1Type           string            `koanf:"parent1.child1.type"`
	Parent1Child1Empty          map[string]string `koanf:"parent1.child1.empty"`
	Parent1Child1Grandchild1IDs []int             `koanf:"parent1.child1.grandchild1.ids"`
	Parent1Child1Grandchild1On  bool              `koanf:"parent1.child1.grandchild1.on"`
}

type customText int

func (c *customText) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))

	switch {
	case strings.HasSuffix(s, "mb"):
		s = strings.TrimSuffix(s, "mb")
	default:
		s = "0"
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*c = customText(v)
	return nil
}

func TestUnmarshal(t *testing.T) {
	assert := assert.New(t)

	// Expected unmarshalled structure.
	real := testStruct{
		Type:  "json",
		Empty: make(map[string]string),
		Parent1: parentStruct{
			Name: "parent1",
			ID:   1234,
			Child1: childStruct{
				Name:  "child1",
				Type:  "json",
				Empty: make(map[string]string),
				Grandchild1: grandchildStruct{
					Ids: []int{1, 2, 3},
					On:  true,
				},
			},
		},
	}

	// Unmarshal and check all parsers.
	for _, c := range cases {
		var (
			k  = koanf.New(delim)
			ts testStruct
		)
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		assert.Nil(k.Unmarshal("", &ts), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")

		// Unmarshal with config.
		ts = testStruct{}
		assert.Nil(k.UnmarshalWithConf("", &ts, koanf.UnmarshalConf{Tag: "koanf"}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")
	}
}

func TestUnmarshalFlat(t *testing.T) {
	assert := assert.New(t)

	// Expected unmarshalled structure.
	real := testStructFlat{
		Type:                        "json",
		Empty:                       make(map[string]string),
		Parent1Name:                 "parent1",
		Parent1ID:                   1234,
		Parent1Child1Name:           "child1",
		Parent1Child1Type:           "json",
		Parent1Child1Empty:          make(map[string]string),
		Parent1Child1Grandchild1IDs: []int{1, 2, 3},
		Parent1Child1Grandchild1On:  true,
	}

	// Unmarshal and check all parsers.
	for _, c := range cases {
		k := koanf.New(delim)
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		ts := testStructFlat{}
		assert.Nil(k.UnmarshalWithConf("", &ts, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1Child1Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")
	}
}

func TestUnmarshalCustomText(t *testing.T) {
	test := struct {
		V1 customText `koanf:"v1"`
		V2 string     `koanf:"v2"`
		V3 customText `koanf:"v3"`
	}{}

	// Test custom unmarshalling of strings via mapstructure's UnmarshalText()
	// methods. customText is an int type that strips of the `mb` suffix and parses
	// the rest into a number.

	k := koanf.New(delim)
	err := k.Load(rawbytes.Provider([]byte(`{"v1": "42mb", "v2": "42mb"}`)), json.Parser())
	assert.NoError(t, err)

	k.Unmarshal("", &test)
	assert.Equal(t, int(test.V1), 42)
	assert.Equal(t, test.V2, "42mb")
	assert.Equal(t, int(test.V3), 0)
}

func TestUnmarshalDecoderConfigTagName(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(rawbytes.Provider([]byte(`{"name": "bob", "other": "alice"}`)), json.Parser()))

	type person struct {
		Label string `mytag:"name" koanf:"other"`
	}

	// A TagName on a user-supplied DecoderConfig must be honoured.
	var p person
	assert.Nil(k.UnmarshalWithConf("", &p, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{TagName: "mytag"},
	}))
	assert.Equal("bob", p.Label)

	// An explicit Tag still wins over the DecoderConfig's TagName.
	var p2 person
	assert.Nil(k.UnmarshalWithConf("", &p2, koanf.UnmarshalConf{
		Tag:           "koanf",
		DecoderConfig: &mapstructure.DecoderConfig{TagName: "mytag"},
	}))
	assert.Equal("alice", p2.Label)

	// A DecoderConfig is not mutated, so it stays reusable across calls.
	shared := &mapstructure.DecoderConfig{}

	var p3 person
	assert.Nil(k.UnmarshalWithConf("", &p3, koanf.UnmarshalConf{Tag: "mytag", DecoderConfig: shared}))
	assert.Equal("bob", p3.Label)
	assert.Equal("", shared.TagName)
	assert.Nil(shared.Result)

	var p4 person
	assert.Nil(k.UnmarshalWithConf("", &p4, koanf.UnmarshalConf{DecoderConfig: shared}))
	assert.Equal("alice", p4.Label)
	assert.Equal("bob", p3.Label)
	assert.Equal("", shared.TagName)
	assert.Nil(shared.Result)
}

func TestTextUnmarshalStringFixed(t *testing.T) {
	defer func() {
		assert.Nil(t, recover())
	}()

	type targetStruct struct {
		LogFormatPointer LogFormatPointer // default should map to json
		LogFormatValue   LogFormatValue   // default should map to json
	}

	target := &targetStruct{"text_custom", "text_custom"}
	before := *target

	var (
		bptr any = &(target.LogFormatPointer)
		cptr any = target.LogFormatValue
	)
	_, ok := (bptr).(encoding.TextMarshaler)
	assert.True(t, ok)

	_, ok = (cptr).(encoding.TextMarshaler)
	assert.True(t, ok)

	k := koanf.New(".")
	k.Load(structs.Provider(target, "koanf"), nil)

	k.Load(env.Provider(".", env.Opt{
		TransformFunc: func(k string, v string) (string, any) {
			return strings.ReplaceAll(strings.ToLower(k), "_", "."), v
		},
	}), nil)

	// default values
	err := k.Unmarshal("", &target)
	assert.NoError(t, err)
	assert.Equal(t, &before, target)
}

// LogFormatValue is a custom string type that implements the TextUnmarshaler interface
// Additionally it implements the TextMarshaler interface (value receiver)
type LogFormatValue string

// pointer receiver
func (c *LogFormatValue) UnmarshalText(data []byte) error {
	switch strings.ToLower(string(data)) {
	case "", "json":
		*c = "json_custom"
	case "text":
		*c = "text_custom"
	default:
		return fmt.Errorf("invalid log format: %s", string(data))
	}
	return nil
}

// value receiver
func (c LogFormatValue) MarshalText() ([]byte, error) {
	// overcomplicated custom internal string representation
	switch c {
	case "", "json_custom":
		return []byte("json"), nil
	case "text_custom":
		return []byte("text"), nil
	}
	return nil, fmt.Errorf("invalid internal string representation: %q", c)
}

// LogFormatPointer is a custom string type that implements the TextUnmarshaler interface
// Additionally it implements the TextMarshaler interface (pointer receiver)
type LogFormatPointer string

// pointer receiver
func (c *LogFormatPointer) UnmarshalText(data []byte) error {
	switch strings.ToLower(string(data)) {
	case "", "json":
		*c = "json_custom"
	case "text":
		*c = "text_custom"
	default:
		return fmt.Errorf("invalid log format: %s", string(data))
	}
	return nil
}

// also pointer receiver
func (c *LogFormatPointer) MarshalText() ([]byte, error) {
	// overcomplicated custom internal string representation
	switch *c {
	case "", "json_custom":
		return []byte("json"), nil
	case "text_custom":
		return []byte("text"), nil
	}
	return nil, fmt.Errorf("invalid internal string representation: %q", *c)
}

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
