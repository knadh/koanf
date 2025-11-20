package dotenv

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotEnv_Marshal(t *testing.T) {
	de := DotEnv{}
	testCases := []struct {
		name   string
		cfg    map[string]any
		expOut []byte
		err    error
	}{
		{
			name:   "Empty config",
			cfg:    map[string]any{},
			expOut: []byte{},
		},
		{
			name: "Simple key-value pair",
			cfg: map[string]any{
				"key": "value",
			},
			expOut: []byte("key=\"value\""),
		},
		{
			name: "Multiple key-values",
			cfg: map[string]any{
				"key_1": "value_1",
				"key_2": "value_2",
				"key_3": "value_3",
			},
			expOut: []byte("key_1=\"value_1\"\nkey_2=\"value_2\"\nkey_3=\"value_3\""),
		},
		{
			name: "Mixed data types",
			cfg: map[string]any{
				"int_key":   12,
				"bool_key":  true,
				"arr_key":   []int{1, 2, 3, 4},
				"float_key": 10.5,
			},
			expOut: []byte("arr_key=\"[1 2 3 4]\"\nbool_key=\"true\"\nfloat_key=\"10.5\"\nint_key=12"),
		},
		{
			name: "Nested config",
			cfg: map[string]any{
				"map_key": map[string]any{
					"arr_key":  []float64{1.2, 4.3, 5, 6},
					"bool_key": false,
					"inner_map_key": map[any]any{
						0: "zero",
						1: 1.0,
					},
					"int_key": 12,
				},
			},
			expOut: []byte(`map_key="map[arr_key:[1.2 4.3 5 6] bool_key:false inner_map_key:map[0:zero 1:1] int_key:12]"`),
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
		expOut map[string]any
		err    bool
	}{
		{
			name:   "Empty config",
			expOut: map[string]any{},
		},
		{
			name: "Simple key_value",
			cfg:  []byte(`key="value"`),
			expOut: map[string]any{
				"key": "value",
			},
		},
		{
			name: "Multiple key_values",
			cfg:  []byte("key_1=\"value_1\"\nkey_2=\"value_2\""),
			expOut: map[string]any{
				"key_1": "value_1",
				"key_2": "value_2",
			},
		},
		{
			name: "Mixed data types",
			cfg:  []byte("arr=\"[1 2 3 4]\"\nbool=\"true\"\nfloat=\"12.5\"\nint=\"32\"\n"),
			expOut: map[string]any{
				"arr":   "[1 2 3 4]",
				"bool":  "true",
				"float": "12.5",
				"int":   "32",
			},
		},
		{
			name: "Nested config",
			cfg:  []byte(`map_key="map[arr_key:[1 4 5 6] bool_key:false inner_map_key:map[0:zero 1:1] int_key:12]"`),
			expOut: map[string]any{
				"map_key": "map[arr_key:[1 4 5 6] bool_key:false inner_map_key:map[0:zero 1:1] int_key:12]",
			},
		},
		{
			name: "Missing quotation mark",
			cfg:  []byte(`key="value`),
			err:  true,
		},
		{
			name: "Missing value",
			cfg:  []byte(`key=`),
			expOut: map[string]any{
				"key": "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outMap, err := de.Unmarshal(tc.cfg)
			if tc.err {
				require.NotNil(t, err)
			}
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

func TestParserEnvWithValue(t *testing.T) {
	testCases := []struct {
		name     string
		prefix   string
		delim    string
		key      string
		value    string
		expKey   string
		expValue any
		cb       func(key, value string) (string, any)
	}{
		{
			name:   "Nil cb",
			prefix: "TESTVAR_",
			delim:  ".",
		},
		{
			name:     "Simple cb",
			prefix:   "TESTVAR_",
			delim:    ".",
			key:      "TestKey",
			value:    "TestVal",
			expKey:   "testkey",
			expValue: "TestVal",
			cb: func(key, value string) (string, any) {
				return strings.ToLower(key), value
			},
		},
		{
			name:   "Empty string nil cb",
			prefix: "",
			delim:  ".",
		},
		{
			name:     "Cb is given",
			prefix:   "",
			delim:    ".",
			key:      "test_key",
			value:    "test_val",
			expKey:   "TEST.KEY",
			expValue: "test_val",
			cb: func(key, value string) (string, any) {
				return strings.Replace(strings.ToUpper(key), "_", ".", -1), value
			},
		},
		{
			name:   "Cb is given and changes value",
			prefix: "",
			delim:  ".",
			key:    "test_key",
			value:  `{"foo": "bar"}`,
			expKey: "TEST.KEY",
			expValue: map[string]any{
				"foo": "bar",
			},
			cb: func(key, value string) (string, any) {
				key = strings.Replace(strings.ToUpper(key), "_", ".", -1)

				var v map[string]any
				err := json.Unmarshal([]byte(value), &v)
				if err == nil {
					return key, v
				}
				return key, value
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotProvider := ParserEnvWithValue(tc.prefix, tc.delim, tc.cb)
			if tc.cb == nil {
				assert.Equal(t, tc.prefix, gotProvider.prefix)
				assert.Equal(t, tc.delim, gotProvider.delim)
			}
			if tc.cb != nil {
				k, v := gotProvider.cb(tc.key, tc.value)
				assert.Equal(t, tc.expKey, k)
				assert.Equal(t, tc.expValue, v)
			}
		})
	}
}
