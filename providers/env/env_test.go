package env

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	mockEnviron := func() []string {
		return []string{"TEST_FOO=bar"}
	}

	testCases := []struct {
		name     string
		prefix   string
		delim    string
		key      string
		value    string
		expKey   string
		expValue string
		cb       func(key string) string
		opt      *Opt
		want     *Env
	}{
		{
			name:   "Nil cb",
			prefix: "TESTVAR_",
			delim:  ".",
			want: &Env{
				prefix: "TESTVAR_",
				delim:  ".",
				opt:    defaultOpt,
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
				opt:    defaultOpt,
			},
		},
		{
			name:   "Empty string nil cb",
			prefix: "",
			delim:  ".",
			want: &Env{
				prefix: "",
				delim:  ".",
				opt:    defaultOpt,
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
		{
			name:   "Custom opt",
			prefix: "TEST_",
			delim:  ".",
			opt: &Opt{
				EnvironFunc: mockEnviron,
			},
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
				opt: &Opt{
					EnvironFunc: mockEnviron,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotProvider *Env
			if tc.opt != nil {
				gotProvider = Provider(tc.prefix, tc.delim, tc.cb, tc.opt)
			} else {
				gotProvider = Provider(tc.prefix, tc.delim, tc.cb)
			}

			if tc.cb == nil && tc.opt == nil {
				assert.Equal(t, tc.want, gotProvider)
				return
			}

			if tc.cb != nil {
				k, v := gotProvider.cb(tc.key, tc.value)
				assert.Equal(t, tc.expKey, k)
				assert.Equal(t, tc.expValue, v)
			}
			if tc.opt != nil {
				wantEnv := tc.want.opt.EnvironFunc()
				gotEnv := gotProvider.opt.EnvironFunc()
				slices.Sort(wantEnv)
				slices.Sort(gotEnv)
				if !slices.Equal(wantEnv, gotEnv) {
					assert.Fail(t, "Env vars not equal (omitted from message for security)",
						"Want len: %d\nGot len:%d", len(wantEnv), len(gotEnv))
				}
			}
		})
	}
}

func TestProviderWithValue(t *testing.T) {
	mockEnviron := func() []string {
		return []string{"TEST_FOO=bar"}
	}

	testCases := []struct {
		name   string
		prefix string
		delim  string
		cb     func(key string, value string) (string, interface{})
		opt    *Opt
		want   *Env
	}{
		{
			name:   "Nil cb",
			prefix: "TEST_",
			delim:  ".",
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
				opt:    defaultOpt,
			},
		},
		{
			name:   "Empty string nil cb",
			prefix: "",
			delim:  ".",
			want: &Env{
				prefix: "",
				delim:  ".",
				opt:    defaultOpt,
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
				opt: defaultOpt,
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
				opt: defaultOpt,
			},
		},
		{
			name:   "Custom opt",
			prefix: "TEST_",
			delim:  ".",
			opt: &Opt{
				EnvironFunc: mockEnviron,
			},
			want: &Env{
				prefix: "TEST_",
				delim:  ".",
				opt: &Opt{
					EnvironFunc: mockEnviron,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got *Env
			if tc.opt != nil {
				got = ProviderWithValue(tc.prefix, tc.delim, tc.cb, tc.opt)
			} else {
				got = ProviderWithValue(tc.prefix, tc.delim, tc.cb)
			}
			if tc.cb == nil && tc.opt == nil {
				assert.Equal(t, tc.want, got)
				return
			}

			if tc.cb != nil {
				keyGot, valGot := got.cb("test_key_env_1", "test_val")
				keyWant, valWant := tc.want.cb("test_key_env_1", "test_val")
				assert.Equal(t, tc.prefix, got.prefix)
				assert.Equal(t, tc.delim, got.delim)
				assert.Equal(t, keyWant, keyGot)
				assert.Equal(t, valWant, valGot)
			}
			if tc.opt != nil {
				wantEnv := tc.want.opt.EnvironFunc()
				gotEnv := got.opt.EnvironFunc()
				slices.Sort(wantEnv)
				slices.Sort(gotEnv)
				if !slices.Equal(wantEnv, gotEnv) {
					assert.Fail(t, "Env vars not equal (omitted from message for security)",
						"Want len: %d\nGot len:%d", len(wantEnv), len(gotEnv))
				}
				assert.Equal(t, tc.want.opt.EnvironFunc(), got.opt.EnvironFunc())
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
				opt:   defaultOpt,
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
				opt: defaultOpt,
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
				opt: defaultOpt,
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
				opt:   defaultOpt,
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
				opt: defaultOpt,
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
				opt:   defaultOpt,
			},
		},
		{
			name:     "Environ func provided",
			key:      "TEST_KEY",
			value:    "TEST_VAL",
			expKey:   "TEST_OVERRIDE_KEY",
			expValue: "TEST_OVERRIDE_VAL",
			env: &Env{
				delim: ".",
				opt: &Opt{
					EnvironFunc: func() []string {
						return []string{"TEST_OVERRIDE_KEY=TEST_OVERRIDE_VAL"}
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)

			envs, err := tc.env.Read()
			assert.Nil(t, err)
			v, ok := envs[tc.expKey]
			assert.True(t, ok)
			assert.Equal(t, tc.expValue, v)
		})
	}
}
