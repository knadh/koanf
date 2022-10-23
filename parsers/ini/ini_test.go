package ini

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestINIUnmarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		keys   []string
		values []interface{}
		isErr  bool
	}{
		{
			name:  "Empty INI",
			input: []byte(``),
		},
		{
			name: "Valid INI - empty section",
			input: []byte(`
				;comment1
				; comment2

				app_mode = development
				n = 81
				`),
			keys:   []string{"app_mode", "n"},
			values: []interface{}{"development", 81},
		},
		{
			name: "Invalid INI - missing square bracket",
			input: []byte(`
				;comment

				app_mode = development

				[sites
				opensource = github
				`),
			isErr: true,
		},
		{
			name: "Complex INI - empty section, all types",
			input: []byte(`
								;comment

								boolean = true
								color = blue
								number = 335
								quote = "No "Us" in this"
							`),
			keys: []string{"boolean", "color", "number", "quote"},
			values: []interface{}{
				true,
				"blue",
				335,
				"\"No \"Us\" in this\"",
			},
		},
		{
			name: "Invalid INI - missing '='",
			input: []byte(`
								;comment

								[http]
								port=8080
								username=httpuser
								[https]
								port 8043
								username=httpsuser
							`),
			isErr: true,
		},
	}
	p := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := p.Unmarshal(tc.input)
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

func TestINIMarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]interface{}
		output []byte
		isErr  bool
	}{
		{
			name:   "Empty INI",
			input:  map[string]interface{}{},
			output: []byte(``),
		},
		{
			name: "Valid INI",
			input: map[string]interface{}{
				"app_mode": "development",
			},
			output: []byte(`app_mode = development
`),
		},
	}

	p := Parser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := p.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}
