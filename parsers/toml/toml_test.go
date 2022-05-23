package toml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTOML_Unmarshal(t *testing.T) {

	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []interface{}
		isErr  bool
	}{
		{
			name:  "Empty TOML",
			input: []byte(``),
		},
		{
			name: "Valid TOML",
			input: []byte(`key = "val"
			name = "test"
			number = 2
			`),
			keys:   []string{"key", "name", "number"},
			values: []interface{}{"val", "test", int64(2)},
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
			keys: []string{"array", "boolean", "color", "null", "number", "object", "string"},
			values: []interface{}{[]interface{}{int64(1), int64(2), int64(3)},
				true,
				"gold",
				nil,
				int64(123),
				map[string]interface{}{"a": "b", "c": "d"},
				"Hello World"},
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
				for i, k := range tc.keys {
					v := out[k]
					assert.Equal(t, tc.values[i], v)
				}
			}
		})
	}
}

func TestTOML_Marshal(t *testing.T) {

	testCases := []struct {
		name   string
		input  map[string]interface{}
		output []byte
		isErr  bool
	}{
		{
			name:   "Empty TOML",
			input:  map[string]interface{}{},
			output: []byte(nil),
		},
		{
			name: "Valid TOML",
			input: map[string]interface{}{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			output: []byte(`key = "val"
name = "test"
number = 2.0
`),
		},
		{
			name: "Complex TOML - All types",
			input: map[string]interface{}{
				"array":   []interface{}{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "gold",
				"number":  123,
				"object":  map[string]interface{}{"a": "b", "c": "d"},
				"string":  "Hello World",
			},
			output: []byte(`array = [1,2,3,4,5]
boolean = true
color = "gold"
number = 123
string = "Hello World"

[object]
  a = "b"
  c = "d"
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
