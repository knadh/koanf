package toml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTOML_Unmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		output map[string]any
		isErr  bool
	}{
		{
			name:   "Empty TOML",
			input:  []byte(``),
			output: map[string]any(nil),
		},
		{
			name: "Valid TOML",
			input: []byte(`key = "val"
			name = "test"
			number = 2
			`),
			output: map[string]any{
				"key":    "val",
				"name":   "test",
				"number": int64(2),
			},
		},
		{
			name:  "Invalid TOML - missing end quotes",
			input: []byte(`key = "val`),
			isErr: true,
		},
		{
			name: "Complex TOML - All types",
			input: []byte(`array = [ 1, 2, 3 ]
					boolean = true
					color = "gold"
					number = 123
					string = "Hello World"
					[object]
					a = "b"
					c = "d"`),
			output: map[string]any{
				"array":   []any{int64(1), int64(2), int64(3)},
				"boolean": true,
				"color":   "gold",
				"number":  int64(123),
				"object":  map[string]any{"a": "b", "c": "d"},
				"string":  "Hello World",
			},
		},
		{
			name:  "Invalid TOML - missing equal",
			input: []byte(`key "val"`),
			isErr: true,
		},
	}

	tp := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tp.Unmarshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}

func TestTOML_Marshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]any
		output []byte
		isErr  bool
	}{
		{
			name:   "Empty TOML",
			input:  map[string]any{},
			output: []byte(nil),
		},
		{
			name: "Valid TOML",
			input: map[string]any{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			output: []byte(`key = 'val'
name = 'test'
number = 2.0
`),
		},
		{
			name: "Complex TOML - All types",
			input: map[string]any{
				"array":   []any{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "gold",
				"number":  123,
				"object":  map[string]any{"a": "b", "c": "d"},
				"string":  "Hello World",
			},
			output: []byte(`array = [1, 2, 3, 4, 5]
boolean = true
color = 'gold'
number = 123
string = 'Hello World'

[object]
a = 'b'
c = 'd'
`),
		},
	}

	tp := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tp.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}
