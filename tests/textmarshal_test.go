package koanf_test

import (
	"encoding"
	"fmt"
	"strings"
	"testing"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

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
