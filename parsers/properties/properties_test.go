package properties

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProperties_Unmarshal(t *testing.T) {
	testCases := []struct {
		name	string
		input	[]byte
		keys	[]string
		values	[]interface{}
		isErr	bool
	}{
		{
			name:	"Empty Properties",
			input:	[]byte(``),
		},
		{
			name:	"Valid Properties",
			input:	[]byte(`
				key=val
				"name" = "test"
				number:2
				amount 4.53
				`),
			keys:	[]string{"key", "\"name\"", "number", "amount"},
			values:	[]interface{}{"val", "\"test\"", 2, 4.53},
		},
	}

	jp := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := jp.Unmarshal(tc.input)
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

func TestProperties_Marshal(t *testing.T) {
	testCases := []struct {
		name	string
		input	map[string]interface{}
		output	[]byte
		isErr	bool
	}{
		{
			name:	"Empty Properties",
			input:	map[string]interface{}{},
			output:	[]byte(``),
		},
		{
			name:	"Complex valid Properties - all types",
			input:	map[string]interface{}{
				"boolean":	true,
				"color":	"gold",
				"number":	123,
				"string":	"Hello World",
			},
			output: []byte("boolean = true\ncolor = gold\nnumber = 123\nstring = Hello World\n"),
		},
	}

	jp := Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := jp.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}
