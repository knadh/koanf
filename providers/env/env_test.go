package env

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProvider(t *testing.T) {

	testCases := []struct {
		name    string
		prefix  string
		delim   string
		cb      func(key string, value string) (string, interface{})
		cbInput func(key string) string
		want    *Env
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
			name:   "Empty string nil cb",
			prefix: "",
			delim:  ".",
			want: &Env{
				prefix: "",
				delim:  ".",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Provider(tc.prefix, tc.delim, tc.cbInput)
			assert.Equal(t, tc.want, got)
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
				key = strings.ToLower(key)
				key = strings.TrimPrefix(key, "test_")
				key = strings.Replace(key, "_", ".", -1)
				return key, value
			},
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
				cb: func(key string, value string) (string, interface{}) {
					key = strings.ToLower(key)
					key = strings.TrimPrefix(key, "test_")
					key = strings.Replace(key, "_", ".", -1)
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
				kGot, vGot := got.cb("test_key_env_1", "test_val")
				kTc, vTc := tc.want.cb("test_key_env_1", "test_val")
				assert.Equal(t, tc.prefix, got.prefix)
				assert.Equal(t, tc.delim, got.delim)
				assert.Equal(t, kTc, kGot)
				assert.Equal(t, vTc, vGot)
				assert.NotNil(t, got.cb)
			}
		})
	}
}
