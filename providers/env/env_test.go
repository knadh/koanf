package env

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	}{
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
			k, v := gotProvider.cb(tc.key, tc.value)
			assert.Equal(t, tc.expKey, k)
			assert.Equal(t, tc.expValue, v)
		})
	}
}

func TestProviderWithValue(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		delim  string
		cb     func(key string, value string) (string, interface{})
		want   *Env
	}{
		{
			name:   "Custom cb function",
			prefix: "TEST_",
			delim:  ".",
			cb: func(key string, value string) (string, interface{}) {
				key = strings.Replace(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".", -1)
				return key, value
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: os.Environ(),
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
			keyGot, valGot := got.cb("test_key_env_1", "test_val")
			keyWant, valWant := tc.want.cb("test_key_env_1", "test_val")
			assert.Equal(t, tc.prefix, got.prefix)
			assert.Equal(t, tc.delim, got.delim)
			assert.Equal(t, keyWant, keyGot)
			assert.Equal(t, valWant, valGot)
		})
	}
}

func TestProviderWithOptions(t *testing.T) {
	testCases := []struct {
		name    string
		options []Option
		want    *Env
	}{
		{
			name: "Nil cb",
			options: []Option{
				WithPrefix("TEST_"),
				WithDelimiter("."),
				WithEnviron([]string{"FOO=BAR"}),
				WithCallback(nil),
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: []string{"FOO=BAR"},
			},
		},
		{
			name: "Empty prefix nil cb",
			options: []Option{
				WithPrefix(""),
				WithDelimiter("."),
				WithEnviron([]string{"FOO=BAR"}),
				WithCallback(nil),
			},
			want: &Env{
				prefix:  "",
				delim:   ".",
				environ: []string{"FOO=BAR"},
			},
		},
		{
			name: "Return the same key-value pair in cb",
			options: []Option{
				WithPrefix("TEST_"),
				WithDelimiter("."),
				WithEnviron([]string{"FOO=BAR"}),
				WithCallback(func(key string, value string) (string, interface{}) {
					return key, value
				}),
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: []string{"FOO=BAR"},
				cb: func(key string, value string) (string, interface{}) {
					return key, value
				},
			},
		},
		{
			name: "Custom cb function",
			options: []Option{
				WithPrefix("TEST_"),
				WithDelimiter("."),
				WithEnviron([]string{"FOO=BAR"}),
				WithCallback(func(key string, value string) (string, interface{}) {
					key = strings.Replace(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".", -1)
					return key, value
				}),
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: []string{"FOO=BAR"},
				cb: func(key string, value string) (string, interface{}) {
					key = strings.Replace(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".", -1)
					return key, value
				},
			},
		},
		{
			name: "with custom environment slice",
			options: []Option{
				WithPrefix("TEST_"),
				WithDelimiter("."),
				WithEnviron([]string{"FOO=BAR"}),
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: []string{"FOO=BAR"},
			},
		},
		{
			name: "with custom environment map",
			options: []Option{
				WithPrefix("TEST_"),
				WithDelimiter("."),
				WithEnvironMap(map[string]string{
					"FOO": "BAR",
				}),
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: []string{"FOO=BAR"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ProviderWithOptions(tc.options...)
			if got.cb == nil {
				assert.Equal(t, tc.want, got)
			} else {
				keyGot, valGot := got.cb("test_key_env_1", "test_val")
				keyWant, valWant := tc.want.cb("test_key_env_1", "test_val")
				assert.Equal(t, keyWant, keyGot)
				assert.Equal(t, valWant, valGot)
			}
		})
	}
}

func TestRead(t *testing.T) {
	testCases := []struct {
		name     string
		expKey   string
		expValue string
		env      *Env
	}{
		{
			name:     "No cb",
			expKey:   "TEST_KEY",
			expValue: "TEST_VAL",
			env: &Env{
				delim: ".",
				environ: []string{
					"TEST_KEY=TEST_VAL",
				},
			},
		},
		{
			name:     "cb given",
			expKey:   "test.key",
			expValue: "TEST_VAL",
			env: &Env{
				delim: "_",
				environ: []string{
					"TEST_KEY=TEST_VAL",
				},
				cb: func(key string, value string) (string, interface{}) {
					return strings.Replace(strings.ToLower(key), "_", ".", -1), value
				},
			},
		},
		{
			name:     "No cb - prefix given",
			expKey:   "test.key",
			expValue: "TEST_VAL",
			env: &Env{
				prefix: "TEST",
				delim:  "/",
				environ: []string{
					"TEST_KEY=TEST_VAL",
				},
				cb: func(key string, value string) (string, interface{}) {
					return strings.Replace(strings.ToLower(key), "_", ".", -1), value
				},
			},
		},
		{
			name:     "Path value",
			expKey:   "TEST_DIR",
			expValue: "/test/dir/file",
			env: &Env{
				environ: []string{
					"TEST_DIR=/test/dir/file",
				},
				delim: ".",
			},
		},
		{
			name:     "Replace value with underscore",
			expKey:   "TEST_DIR",
			expValue: "_test_dir_file",
			env: &Env{
				delim: ".",
				environ: []string{
					"TEST_DIR=/test/dir/file",
				},
				cb: func(key string, value string) (string, interface{}) {
					return key, strings.Replace(strings.ToLower(value), "/", "_", -1)
				},
			},
		},
		{
			name:     "Empty value",
			expKey:   "KEY",
			expValue: "",
			env: &Env{
				environ: []string{
					"KEY=",
				},
				delim: ".",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envs, err := tc.env.Read()
			assert.Nil(t, err)
			v, ok := envs[tc.expKey]
			assert.True(t, ok)
			assert.Equal(t, tc.expValue, v)
		})
	}
}
