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
			input:  []byte(`key "val" ; name "test" ; number 2.0`),
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
				array 1.0 2.0 3.0
				boolean true
				color "gold"
				"null" null
				number 123
				object a="b" c="d" e=2.7 f=true
				string "Hello World"
			`),
			keys: []string{"array", "boolean", "color", "null", "number", "object", "string"},
			values: []interface{}{[]interface{}{1.0, 2.0, 3.0},
				true,
				"gold",
				nil,
				int64(123),
				map[string]interface{}{"a": "b", "c": "d", "e": 2.7, "f": true},
				"Hello World"},
		},
		{
			name:  "Invalid KDL - missing value",
			input: []byte(`node1 boolean=`),
			isErr: true,
		},
		{
			name: "Complex KDL - Nested map",
			input: []byte(`key "value"
					"1" "skipped"
					map key="skipped" key="value"
					nested_map {
						map key="value" 17 {
							list "item1" "item2" "item3"
							mixup "y"=1 2 3 4
							first "first"=1 2 3 4
							child "test"=1 2 3 4 { "y" 5 ; "d" 6 ; }
						}
					}
				`),
			keys: []string{"key", "1", "map", "nested_map"},
			values: []interface{}{
				"value",
				"skipped",
				map[string]interface{}{
					"key": "value",
				},
				map[string]interface{}{
					"map": map[string]interface{}{
						"0":   int64(17),
						"key": "value",
						"list": []interface{}{
							"item1",
							"item2",
							"item3",
						},
						"mixup": map[string]interface{}{
							"y": int64(1),
							"0": int64(2),
							"1": int64(3),
							"2": int64(4),
						},
						"first": map[string]interface{}{
							"first": int64(1),
							"0":     int64(2),
							"1":     int64(3),
							"2":     int64(4),
						},
						"child": map[string]interface{}{
							"test": int64(1),
							"0":    int64(2),
							"1":    int64(3),
							"2":    int64(4),
							"y":    int64(5),
							"d":    int64(6),
						},
					},
				},
			},
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
		name              string
		input             map[string]interface{}
		stringifiedOutput string
		isErr             bool
	}{
		{
			name:              "Empty KDL",
			input:             map[string]interface{}{},
			stringifiedOutput: ``,
		},
		{
			name: "Valid KDL",
			input: map[string]interface{}{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			stringifiedOutput: `key "val"
name "test"
number 2.0
`,
		},
		{
			name: "Complex KDL - Different types",
			input: map[string]interface{}{
				"null":    nil,
				"boolean": true,
				"color":   "gold",
				"number":  int64(123),
				"string":  "Hello World",
				// "array":   []interface{}{1, 2, 3, 4, 5}, // https://github.com/sblinch/kdl-go/issues/3
				"object": map[string]interface{}{"a": "b", "c": "d"},
			},
			stringifiedOutput: `boolean true
color "gold"
number 123
string "Hello World"
object a="b" c="d"
null null
`,
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
				assert.Equal(t, tc.stringifiedOutput, string(out))
			}
		})
	}
}
