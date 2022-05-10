package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAML_Unmarshal(t *testing.T) {

	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []interface{}
		isErr  bool
	}{
		{
			name:  "Empty YAML",
			input: []byte(`{}`),
		},
		{
			name: "Valid YAML",
			input: []byte(`key: val
name: test
number: 2`),
			keys:   []string{"key", "name", "number"},
			values: []interface{}{"val", "test", 2},
		},
		{
			name: "Invalid YAML - wrong intendation",
			input: []byte(`key: val
			name: test
			number: 2`),
			isErr: true,
		},
		{
			name: "Complex YAML - All types",
			input: []byte(`---
array:
- 1
- 2
- 3
boolean: true
color: gold
'null':
number: 123
object:
  a: b
  c: d
string: Hello World`),
			keys: []string{"array", "boolean", "color", "null", "number", "object", "string"},
			values: []interface{}{[]interface{}{1, 2, 3},
				true,
				"gold",
				nil,
				123,
				map[interface{}]interface{}{"a": "b", "c": "d"},
				"Hello World"},
		},
		{
			name: "Valid YAML - With comments",
			input: []byte(`---
key: #Here is a single-line comment 
- value line 5
#Here is a 
#multi-line comment
- value line 13`),
			keys:   []string{"key"},
			values: []interface{}{[]interface{}{"value line 5", "value line 13"}},
		},
	}

	y := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := y.Unmarshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				for i, k := range tc.keys {
					v := out[k]
					fmt.Println(out[k])
					assert.Equal(t, tc.values[i], v)
				}
			}
		})
	}
}

func TestYAML_Marshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]interface{}
		output []byte
		isErr  bool
	}{
		{
			name:  "Empty YAML",
			input: map[string]interface{}{},
			output: []byte(`{}
`),
		},
		{
			name: "Valid YAML",
			input: map[string]interface{}{
				"key":    "val",
				"name":   "test",
				"number": 2,
			},
			output: []byte(`key: val
name: test
number: 2
`),
		},
		{
			name: "Complex YAML - All types",
			input: map[string]interface{}{
				"array":   []interface{}{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "gold",
				"null":    nil,
				"number":  123,
				"object":  map[string]interface{}{"a": "b", "c": "d"},
				"string":  "Hello World",
			},
			output: []byte(`array:
- 1
- 2
- 3
- 4
- 5
boolean: true
color: gold
"null": null
number: 123
object:
  a: b
  c: d
string: Hello World
`),
		},
	}

	y := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := y.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}
