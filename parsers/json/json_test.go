package json

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSON_Unmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []interface{}
		isErr  bool
	}{
		{
			name:  "Empty JSON",
			input: []byte(`{}`),
		},
		{
			name: "Valid JSON",
			input: []byte(`{
						"key": "val",
						"name": "test",
						"number": 2
					}`),
			keys:   []string{"key", "name", "number"},
			values: []interface{}{"val", "test", 2.0},
		},
		{
			name: "Invalid JSON - missing curly brace",
			input: []byte(`{
						"key": "val",`),
			isErr: true,
		},
		{
			name: "Complex JSON - All types",
			input: []byte(`{
						  "array": [
							1,
							2,
							3
						  ],
						  "boolean": true,
						  "color": "gold",
						  "null": null,
						  "number": 123,
						  "object": {
							"a": "b",
							"c": "d"
						  },
						  "string": "Hello World"
						}`),
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
			name: "Invalid JSON - missing comma",
			input: []byte(`{
 					 	"boolean": true
  						"number": 123
						}`),
			isErr: true,
		},
		{
			name: "Invalid JSON - Redundant comma",
			input: []byte(`{
  						"number": 123,
						}`),
			isErr: true,
		},
	}
	j := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := j.Unmarshal(tc.input)
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

func TestJSON_Marshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]interface{}
		output []byte
		isErr  bool
	}{
		{
			name:   "Empty JSON",
			input:  map[string]interface{}{},
			output: []byte(`{}`),
		},
		{
			name: "Valid JSON",
			input: map[string]interface{}{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			output: []byte(`{"key":"val","name":"test","number":2}`),
		},
		{
			name: "Complex JSON - All types",
			input: map[string]interface{}{
				"array":   []interface{}{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "gold",
				"null":    nil,
				"number":  123,
				"object":  map[string]interface{}{"a": "b", "c": "d"},
				"string":  "Hello World",
			},
			output: []byte(`{"array":[1,2,3,4,5],"boolean":true,"color":"gold","null":null,"number":123,"object":{"a":"b","c":"d"},"string":"Hello World"}`),
		},
	}

	j := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := j.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}
