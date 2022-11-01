package dotenv

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotEnv_Marshal(t *testing.T) {
	de := DotEnv{}
	testCases := []struct {
		name   string
		cfg    map[string]interface{}
		expOut []byte
		err    error
	}{
		{
			name:   "Empty config",
			cfg:    map[string]interface{}{},
			expOut: []byte{},
		},
		{
			name: "Simple key-value pair",
			cfg: map[string]interface{}{
				"key": "value",
			},
			expOut: []byte("key=\"value\""),
		},
		{
			name: "Multiple key-values",
			cfg: map[string]interface{}{
				"key-1": "value-1",
				"key-2": "value-2",
				"key-3": "value-3",
			},
			expOut: []byte("key-1=\"value-1\"\nkey-2=\"value-2\"\nkey-3=\"value-3\""),
		},
		{
			name: "Mixed data types",
			cfg: map[string]interface{}{
				"int-key":   12,
				"bool-key":  true,
				"arr-key":   []int{1, 2, 3, 4},
				"float-key": 10.5,
			},
			expOut: []byte("arr-key=\"[1 2 3 4]\"\nbool-key=\"true\"\nfloat-key=\"10.5\"\nint-key=\"12\""),
		},
		{
			name: "Nested config",
			cfg: map[string]interface{}{
				"map-key": map[string]interface{}{
					"arr-key":  []float64{1.2, 4.3, 5, 6},
					"bool-key": false,
					"inner-map-key": map[interface{}]interface{}{
						0: "zero",
						1: 1.0,
					},
					"int-key": 12,
				},
			},
			expOut: []byte(`map-key="map[arr-key:[1.2 4.3 5 6] bool-key:false inner-map-key:map[0:zero 1:1] int-key:12]"`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := de.Marshal(tc.cfg)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, string(tc.expOut), string(out))
		})
	}
}

func TestDotEnv_Unmarshal(t *testing.T) {
	de := DotEnv{}
	testCases := []struct {
		name   string
		cfg    []byte
		expOut map[string]interface{}
		err    error
	}{
		{
			name:   "Empty config",
			expOut: map[string]interface{}{},
		},
		{
			name: "Simple key-value",
			cfg:  []byte(`key="value"`),
			expOut: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "Multiple key-values",
			cfg:  []byte("key-1=\"value-1\"\nkey-2=\"value-2\""),
			expOut: map[string]interface{}{
				"key-1": "value-1",
				"key-2": "value-2",
			},
		},
		{
			name: "Mixed data types",
			cfg:  []byte("arr=\"[1 2 3 4]\"\nbool=\"true\"\nfloat=\"12.5\"\nint=\"32\"\n"),
			expOut: map[string]interface{}{
				"arr":   "[1 2 3 4]",
				"bool":  "true",
				"float": "12.5",
				"int":   "32",
			},
		},
		{
			name: "Nested config",
			cfg:  []byte(`map-key="map[arr-key:[1 4 5 6] bool-key:false inner-map-key:map[0:zero 1:1] int-key:12]"`),
			expOut: map[string]interface{}{
				"map-key": "map[arr-key:[1 4 5 6] bool-key:false inner-map-key:map[0:zero 1:1] int-key:12]",
			},
		},
		{
			name: "Missing quotation mark",
			cfg:  []byte(`key="value`),
			expOut: map[string]interface{}{
				"key": `"value`,
			},
		},
		{
			name: "Missing value",
			cfg:  []byte(`key=`),
			expOut: map[string]interface{}{
				"key": "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outMap, err := de.Unmarshal(tc.cfg)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expOut, outMap)
		})
	}
}

func TestCompareToEnvProvider(t *testing.T) {

	testCases := []struct {
		name     string
		prefix   string
		delim    string
		key      string
		value    string
		expKey   string
		expValue string
		cb       func(key string) string
		want     *DotEnv
	}{
		{
			name:   "Nil cb",
			prefix: "TESTVAR_",
			delim:  ".",
			want: &DotEnv{
				prefix: "TESTVAR_",
				delim:  ".",
			},
		},
		{
			name:     "Simple cb",
			prefix:   "TESTVAR_",
			delim:    ".",
			key:      "TestKey",
			value:    "TestVal",
			expKey:   "testkey",
			expValue: "TestVal",
			cb: func(key string) string {
				return strings.ToLower(key)
			},
			want: &DotEnv{
				prefix: "TESTVAR_",
				delim:  ".",
			},
		},
		{
			name:   "Empty string nil cb",
			prefix: "",
			delim:  ".",
			want: &DotEnv{
				prefix: "",
				delim:  ".",
			},
		},
		{
			name:     "Cb is given",
			prefix:   "",
			delim:    ".",
			key:      "test_key",
			value:    "test_val",
			expKey:   "TEST.KEY",
			expValue: "test_val",
			cb: func(key string) string {
				return strings.Replace(strings.ToUpper(key), "_", ".", -1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotProvider := ParserEnv(tc.prefix, tc.delim, tc.cb)
			if tc.cb == nil {
				assert.Equal(t, tc.prefix, gotProvider.prefix)
				assert.Equal(t, tc.delim, gotProvider.delim)
				// do not compare cb or reverseCB
			}
			if tc.cb != nil {
				k, v := gotProvider.cb(tc.key, tc.value)
				assert.Equal(t, tc.expKey, k)
				assert.Equal(t, tc.expValue, v)
			}
		})
	}
}
