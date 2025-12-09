package hjson

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHJSON_Unmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []any
		isErr  bool
	}{
		{
			name:  "Empty HJSON",
			input: []byte(`{}`),
		},
		{
			name: "Valid HJSON",
			input: []byte(`{
					key: val
					name: test
					number: 2
				}`),
			keys:   []string{"key", "name", "number"},
			values: []any{"val", "test", 2.0},
		},
		{
			name: "Commented HJSON",
			input: []byte(`{
					# hash style comments
					# (because it's just one character)

					// line style comments
					// (because it's like C/JavaScript/...)

					/* block style comments because
					   it allows you to comment out a block */

					key: v1
					name: Comments
					number: 3
				}`),
			keys:   []string{"key", "name", "number"},
			values: []any{"v1", "Comments", 3.0},
		},
		{
			name: "Quoted strings HJSON",
			input: []byte(`{
					JSON: "a string"

					HJSON: a string

					#notice, no escape necessary:
					RegEx: \s+
				}`),
			keys:   []string{"JSON", "HJSON", "RegEx"},
			values: []any{"a string", "a string", `\s+`},
		},
		{
			name: "Multiline strings HJSON",
			input: []byte(`{
				md:
					'''
					First line.
					Second line.
					  This line is indented by two spaces.
					'''
				}`),
			keys:   []string{"md"},
			values: []any{"First line.\nSecond line.\n  This line is indented by two spaces."},
		},
		{
			name: "Punctuators HJSON",
			input: []byte(`{
				"key name": "{ sample }"
				"()": " sample at the start/end "
				this: is OK though: {}[],:
				}`),
			keys:   []string{"key name", "()", "this"},
			values: []any{"{ sample }", " sample at the start/end ", "is OK though: {}[],:"},
		},
		{
			name: "Invalid HJSON - missing curly brace",
			input: []byte(`{
				key: val,`),
			isErr: true,
		},
		{
			name: "Comma HJSON",
			input: []byte(`{
					key: a,
					key_n: 3
					key3: b,
				}`),
			keys:   []string{"key", "key_n", "key3"},
			values: []any{"a,", 3.0, "b,"},
		},
		{
			name: "One quote in key - HJSON",
			input: []byte(`{
					"key: a
					"key2": b
				}`),
			isErr: true,
		},
		{
			name: "Comment without marks HJSON",
			input: []byte(`{
					Wrong commentary without #, // or /**/

					key: 1
				}`),
			isErr: true,
		},
		{
			name: "Wrong comment mark HJSON",
			input: []byte(`{
					$ Wrong comment mark

					key: a
				}`),
			isErr: true,
		},
		{
			name: "Wrong line style comment HJSON",
			input: []byte(`{
					/ Wrong comment - / instead of //

					key: a
				}`),
			isErr: true,
		},
		{
			name: "Wrong multiline style comment HJSON",
			input: []byte(`{
					/* Multiline comment without second mark

					key: b
				}`),
			isErr: true,
		},
		{
			name: "Whitespace in the key - HJSON",
			input: []byte(`{
					key whitespace: a
				}`),
			isErr: true,
		},
		{
			name: "Punctuator (comma) in the key - HJSON",
			input: []byte(`{
					key, comma: a
				}`),
			isErr: true,
		},
		{
			name: "Complex HJSON - all types",
			input: []byte(`{
						# All types

						// and comments

						/* including
						multiline style
						comments */

						array: [
							1,
							2,
							3
						]

						boolean: true,
						color: gold
						null: null
						number: 123
						object: {
							"a": "b",
							"c": "d"
						},
						string: "Hello World"
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
	}
	h := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := h.Unmarshal(tc.input)
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

func TestHJSON_Marshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]any
		output map[string]any
	}{
		{
			name:   "Empty HJSON",
			input:  map[string]any{},
			output: map[string]any{},
		},
		{
			name: "Valid HJSON",
			input: map[string]any{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
			output: map[string]any{
				"key":    "val",
				"name":   "test",
				"number": 2.0,
			},
		},
		{
			name: "Multiline value HJSON",
			input: map[string]any{
				"md": `
					'''
					First line.
					Second line.
					  This line is indented by two spaces.
					'''`,
			},
			output: map[string]any{
				"md": `
					'''
					First line.
					Second line.
					  This line is indented by two spaces.
					'''`,
			},
		},
		{
			name: "Complex HJSON - All types",
			input: map[string]any{
				"array":   []any{1, 2, 3, 4, 5},
				"boolean": true,
				"color":   "red",
				"null":    nil,
				"number":  123,
				"object":  map[string]any{"a": "b", "c": "d"},
				"string":  "Hello HJSON",
			},
			output: map[string]any{
				"array":   []any{1.0, 2.0, 3.0, 4.0, 5.0},
				"boolean": true,
				"color":   "red",
				"null":    nil,
				"number":  123.0,
				"object":  map[string]any{"a": "b", "c": "d"},
				"string":  "Hello HJSON",
			},
		},
	}

	h := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf, err := h.Marshal(tc.input)
			assert.Nil(t, err)
			out, err := h.Unmarshal(buf)
			assert.Nil(t, err)
			for key, val := range out {
				assert.Equal(t, tc.output[key], val)
			}
		})
	}
}
