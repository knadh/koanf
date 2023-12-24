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
			input:  []byte(`key "val" ; name "test" ; number 2`),
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
				array 1 2 3
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
				123.0,
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
					map "skipped"
					map key="skipped" key="value"
					nested_map {
						map key="value" 17 {
							list "item1" "item2" "item3"
							mixup ""=1 2 3 4
							first "first"=1 2 3 4
							child ""=1 2 3 4 { "" 5 ; "" 6 ; }
						}
					}
				`),
			keys: []string{"key", "map", "nested_map"},
			values: []interface{}{
				"value",
				map[string]interface{}{
					"key": "value",
				},
				map[string]interface{}{
					"map": map[string]interface{}{
						"":    17,
						"key": "value",
						"list": []interface{}{
							"item1",
							"item2",
							"item3",
						},
						"mixup": map[string]interface{}{
							"": []interface{}{
								2,
								3,
								4,
							},
						},
						"first": map[string]interface{}{
							"first": 1,
							"": []interface{}{
								2,
								3,
								4,
							},
						},
						"child": map[string]interface{}{
							"": 6,
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
