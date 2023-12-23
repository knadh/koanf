package kdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKDL_Unmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []interface{}
		isErr  bool
	}{
		{
			name:  "Empty KDL",
			input: []byte(``),
		},
		{
			name:   "Valid KDL",
			input:  []byte(`node1 key="val" name="test" number=2`),
			keys:   []string{"key", "name", "number"},
			values: []interface{}{"val", "test", 2.0},
		},
		{
			name:  "Invalid KDL - syntax error",
			input: []byte(`node1 key="val`),
			isErr: true,
		},
		{
			name: "Complex KDL - Different types",
			input: []byte(`
				node1 array=[1, 2, 3]
				node2 boolean=true
				node3 color="gold"
				node4 null=null
				node5 number=123
				node6 object={a="b" c="d"}
				node7 string="Hello World"
			`),
			keys: []string{"array", "boolean", "color", "null", "number", "object", "string"},
			values: []interface{}{[]interface{}{1.0, 2.0, 3.0},
				true,
				"gold",
				nil,
				123.0,
				map[string]interface{}{"a": "b", "c": "d"},
				"Hello World"},
		},
		{
			name:  "Invalid KDL - missing value",
			input: []byte(`node1 boolean=`),
			isErr: true,
		},
	}

	k := Parser() // Assuming Parser() is implemented for KDL

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := k.Unmarshal(tc.input)
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

func TestKDL_Marshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]interface{}
		output []byte
		isErr  bool
	}{
		{
			name:   "Empty KDL",
			input:  map[string]interface{}{},
			output: []byte(``),
		},
		{
			name: "Valid KDL",
			input: map[string]interface{}{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			output: []byte(`node key="val" name="test" number=2`),
		},
		{
			name: "Complex KDL - Different types",
			input: map[string]interface{}{
				"array":   []interface{}{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "gold",
				"null":    nil,
				"number":  123,
				"object":  map[string]interface{}{"a": "b", "c": "d"},
				"string":  "Hello World",
			},
			output: []byte(`node array=[1,2,3,4,5] boolean=true color="gold" null=null number=123 object={a="b" c="d"} string="Hello World"`),
		},
	}

	k := Parser() // Assuming Parser() is implemented for KDL

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := k.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)

			}
		})
	}
}
