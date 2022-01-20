package env

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestProvider(t *testing.T) {

	testCases := []struct {
		name     string
		prefix   string
		delim    string
		key      string
		value    string
		expKey   string
		expValue string
		cb       func(key string) string
		want     *Env
	}{
		{
			name:   "Nil cb",
			prefix: "TESTVAR_",
			delim:  ".",
			want: &Env{
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
			want: &Env{
				prefix: "TESTVAR_",
				delim:  ".",
			},
		},
		{
			name:   "Empty string nil cb",
			prefix: "",
			delim:  ".",
			want: &Env{
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
			gotProvider := Provider(tc.prefix, tc.delim, tc.cb)
			if tc.cb == nil {
				assert.Equal(t, tc.want, gotProvider)
			}
			if tc.cb != nil {
				k, v := gotProvider.cb(tc.key, tc.value)
				assert.Equal(t, tc.expKey, k)
				assert.Equal(t, tc.expValue, v)
			}
		})
	}
}

func TestProviderWithValue(t *testing.T) {
	testCases := []struct {
		name        string
		prefix      string
		delim       string
		cb          func(key string, value string) (string, interface{})
		nilCallback bool
		want        *Env
	}{
		{
			name:        "Nil cb",
			prefix:      "TEST_",
			delim:       ".",
			nilCallback: true,
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
			},
		},
		{
			name:        "Empty string nil cb",
			prefix:      "",
			delim:       ".",
			nilCallback: true,
			want: &Env{
				prefix: "",
				delim:  ".",
			},
		},
		{
			name:   "Return the same key-value pair in cb",
			prefix: "TEST_",
			delim:  ".",
			cb: func(key string, value string) (string, interface{}) {
				return key, value
			},
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
				cb: func(key string, value string) (string, interface{}) {
					return key, value
				},
			},
		},
		{
			name:   "Custom cb function",
			prefix: "TEST_",
			delim:  ".",
			cb: func(key string, value string) (string, interface{}) {
				key = strings.Replace(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".", -1)
				return key, value
			},
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
				cb: func(key string, value string) (string, interface{}) {
					key = strings.Replace(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".", -1)
					return key, value
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ProviderWithValue(tc.prefix, tc.delim, tc.cb)
			if tc.nilCallback {
				assert.Equal(t, tc.want, got)
			} else {
				keyGot, valGot := got.cb("test_key_env_1", "test_val")
				keyWant, valWant := tc.want.cb("test_key_env_1", "test_val")
				assert.Equal(t, tc.prefix, got.prefix)
				assert.Equal(t, tc.delim, got.delim)
				assert.Equal(t, keyWant, keyGot)
				assert.Equal(t, valWant, valGot)
			}
		})
	}
}

func TestRead(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		value    string
		expKey   string
		expValue string
		env      *Env
	}{
		{
			name:     "No cb",
			key:      "TEST_KEY",
			value:    "TEST_VAL",
			expKey:   "TEST_KEY",
			expValue: "TEST_VAL",
			env: &Env{
				delim: ".",
			},
		},
		{
			name:     "cb given",
			key:      "TEST_KEY",
			value:    "TEST_VAL",
			expKey:   "test.key",
			expValue: "TEST_VAL",
			env: &Env{
				delim: "_",
				cb: func(key string, value string) (string, interface{}) {
					return strings.Replace(strings.ToLower(key), "_", ".", -1), value
				},
			},
		},
		{
			name:     "No cb - prefix given",
			key:      "TEST_KEY",
			value:    "TEST_VAL",
			expKey:   "test.key",
			expValue: "TEST_VAL",
			env: &Env{
				prefix: "TEST",
				delim:  "/",
				cb: func(key string, value string) (string, interface{}) {
					return strings.Replace(strings.ToLower(key), "_", ".", -1), value
				},
			},
		},
		{
			name:     "Path value",
			key:      "TEST_DIR",
			value:    "/test/dir/file",
			expKey:   "TEST_DIR",
			expValue: "/test/dir/file",
			env: &Env{
				delim: ".",
			},
		},
		{
			name:     "Replace value with underscore",
			key:      "TEST_DIR",
			value:    "/test/dir/file",
			expKey:   "TEST_DIR",
			expValue: "_test_dir_file",
			env: &Env{
				delim: ".",
				cb: func(key string, value string) (string, interface{}) {
					return key, strings.Replace(strings.ToLower(value), "/", "_", -1)
				},
			},
		},
		{
			name:     "Empty value",
			key:      "KEY",
			value:    "",
			expKey:   "KEY",
			expValue: "",
			env: &Env{
				delim: ".",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := os.Setenv(tc.key, tc.value)
			assert.Nil(t, err)
			defer os.Unsetenv(tc.key)

			envs, err := tc.env.Read()
			assert.Nil(t, err)
			v, ok := envs[tc.expKey]
			assert.True(t, ok)
			assert.Equal(t, tc.expValue, v)
		})
	}
}
