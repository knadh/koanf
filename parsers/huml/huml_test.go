package huml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHUML_Unmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		output map[string]any
		isErr  bool
	}{
		{
			name:  "Empty HUML",
			input: []byte(``),
			isErr: true, // HUML considers empty document as undefined
		},
		{
			name:  "Invalid HUML syntax",
			input: []byte(`invalid: syntax: error`),
			isErr: true,
		},
		{
			name: "Simple key-value pairs",
			input: []byte(`name: "test"
port: 8080
debug: true`),
			output: map[string]any{
				"name":  "test",
				"port":  int64(8080),
				"debug": true,
			},
		},
		{
			name: "Array values",
			input: []byte(`tags:: "web", "api", "go"
ports:: 80, 443, 8080`),
			output: map[string]any{
				"tags":  []any{"web", "api", "go"},
				"ports": []any{int64(80), int64(443), int64(8080)},
			},
		},
		{
			name: "Nested objects",
			input: []byte(`database::
  host: "localhost"
  port: 5432
  ssl: false`),
			output: map[string]any{
				"database": map[string]any{
					"host": "localhost",
					"port": int64(5432),
					"ssl":  false,
				},
			},
		},
		{
			name: "Complex nested structure",
			input: []byte(`server::
  http::
    port: 8080
    timeout: 30
  https::
    port: 8443
    enabled: true`),
			output: map[string]any{
				"server": map[string]any{
					"http": map[string]any{
						"port":    int64(8080),
						"timeout": int64(30),
					},
					"https": map[string]any{
						"port":    int64(8443),
						"enabled": true,
					},
				},
			},
		},
		{
			name: "Mixed types",
			input: []byte(`string_val: "hello"
int_val: 42
float_val: 3.14
bool_val: true
array_val:: 1, 2, 3`),
			output: map[string]any{
				"string_val": "hello",
				"int_val":    int64(42),
				"float_val":  3.14,
				"bool_val":   true,
				"array_val":  []any{int64(1), int64(2), int64(3)},
			},
		},
	}

	hp := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := hp.Unmarshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}

func TestHUML_Marshal(t *testing.T) {
	testCases := []struct {
		name  string
		input map[string]any
		isErr bool
	}{
		{
			name:  "Empty map",
			input: map[string]any{},
		},
		{
			name: "Simple values",
			input: map[string]any{
				"name":  "test",
				"port":  int64(8080),
				"debug": true,
			},
		},
		{
			name: "Complex nested structure",
			input: map[string]any{
				"server": map[string]any{
					"http": map[string]any{
						"port":    int64(8080),
						"timeout": int64(30),
					},
					"https": map[string]any{
						"port":    int64(8443),
						"enabled": true,
					},
				},
				"tags": []any{"web", "api", "go"},
			},
		},
	}

	hp := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := hp.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, out)

				// Test round-trip: marshal then unmarshal should give back the same data
				roundTrip, err := hp.Unmarshal(out)
				assert.Nil(t, err)
				assert.Equal(t, tc.input, roundTrip)
			}
		})
	}
}
