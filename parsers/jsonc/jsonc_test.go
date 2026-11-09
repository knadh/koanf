package jsonc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONC_Unmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []any
		isErr  bool
	}{
		{
			name:  "Empty JSONC",
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
			values: []any{"val", "test", 2.0},
		},
		{
			name: "Line and block comments",
			input: []byte(`{
						// A line comment.
						"key": "val", // A trailing comment.
						/* A block
						   comment. */
						"number": /* inline */ 2
					}`),
			keys:   []string{"key", "number"},
			values: []any{"val", 2.0},
		},
		{
			name: "Trailing commas",
			input: []byte(`{
						"array": [1, 2, 3,],
						"object": {"a": "b",},
					}`),
			keys:   []string{"array", "object"},
			values: []any{[]any{1.0, 2.0, 3.0}, map[string]any{"a": "b"}},
		},
		{
			name: "Comment markers inside strings are preserved",
			input: []byte(`{
						"url": "https://example.com/path",
						"glob": "/* not a comment */"
					}`),
			keys:   []string{"url", "glob"},
			values: []any{"https://example.com/path", "/* not a comment */"},
		},
		{
			name: "Complex JSONC - All types",
			input: []byte(`{
						  // Arrays.
						  "array": [
							1,
							2,
							3,
						  ],
						  "boolean": true,
						  "color": "gold",
						  "null": null,
						  "number": 123,
						  /* Nested objects. */
						  "object": {
							"a": "b",
							"c": "d",
						  },
						  "string": "Hello World",
						}`),
			keys: []string{"array", "boolean", "color", "null", "number", "object", "string"},
			values: []any{[]any{1.0, 2.0, 3.0},
				true,
				"gold",
				nil,
				123.0,
				map[string]any{"a": "b", "c": "d"},
				"Hello World"},
		},
		{
			name: "Invalid JSONC - missing curly brace",
			input: []byte(`{
						"key": "val",`),
			isErr: true,
		},
		{
			name: "Invalid JSONC - missing comma",
			input: []byte(`{
						"boolean": true
						"number": 123
					}`),
			isErr: true,
		},
		{
			name:  "Invalid JSONC - unterminated block comment",
			input: []byte(`{"key": "val"} /* comment`),
			isErr: true,
		},
		{
			name:  "Invalid JSONC - hash comment",
			input: []byte("{\n# comment\n\"key\": \"val\"\n}"),
			isErr: true,
		},
		{
			name:  "Invalid JSONC - unquoted key",
			input: []byte(`{key: "val"}`),
			isErr: true,
		},
		{
			name:  "Invalid JSONC - not an object",
			input: []byte(`[1, 2, 3]`),
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

func TestJSONC_UnmarshalDoesNotModifyInput(t *testing.T) {
	input := []byte(`{
		// A comment.
		"key": "val",
	}`)
	original := string(input)

	_, err := Parser().Unmarshal(input)
	assert.Nil(t, err)
	assert.Equal(t, original, string(input))
}

func TestJSONC_Marshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]any
		output []byte
		isErr  bool
	}{
		{
			name:   "Empty JSONC",
			input:  map[string]any{},
			output: []byte(`{}`),
		},
		{
			name: "Valid JSONC",
			input: map[string]any{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			output: []byte(`{"key":"val","name":"test","number":2}`),
		},
		{
			name: "Complex JSONC - All types",
			input: map[string]any{
				"array":   []any{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "gold",
				"null":    nil,
				"number":  123,
				"object":  map[string]any{"a": "b", "c": "d"},
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
