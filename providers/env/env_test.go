package env

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	mockEnviron := func() []string {
		return []string{"TEST_FOO=bar"}
	}
	xformSimple := func(k, v string) (string, any) {
		return strings.ToLower(k), v
	}
	xformTwo := func(k, v string) (string, any) {
		return strings.ReplaceAll(strings.ToUpper(k), "_", "."), v
	}

	testCases := []struct {
		name     string
		prefix   string
		delim    string
		key      string
		value    string
		expKey   string
		expValue string
		opt      Opt
		want     *Env
	}{
		{
			name:  "Nil TransformFunc",
			delim: ".",
			opt:   Opt{Prefix: "TESTVAR_"},
			want: &Env{
				prefix:  "TESTVAR_",
				delim:   ".",
				environ: os.Environ,
			},
		},
		{
			name:     "Simple TransformFunc",
			delim:    ".",
			key:      "TestKey",
			value:    "TestVal",
			expKey:   "testkey",
			expValue: "TestVal",
			opt: Opt{
				Prefix:        "TESTVAR_",
				TransformFunc: xformSimple,
			},
			want: &Env{
				prefix:  "TESTVAR_",
				delim:   ".",
				environ: os.Environ,
			},
		},
		{
			name:  "Empty options",
			delim: ".",
			opt:   Opt{},
			want: &Env{
				delim:   ".",
				environ: os.Environ,
			},
		},
		{
			name:     "TransformFunc is given, no Prefix",
			delim:    ".",
			key:      "test_key",
			value:    "test_val",
			expKey:   "TEST.KEY",
			expValue: "test_val",
			opt: Opt{
				TransformFunc: xformTwo,
			},
			want: &Env{
				delim:   ".",
				environ: os.Environ,
			},
		},
		{
			name:     "Custom cb function",
			delim:    ".",
			key:      "TEST_KEY",
			value:    "test_val",
			expKey:   "key",
			expValue: "prod_val",
			opt: Opt{
				Prefix: "TEST_",
				TransformFunc: func(key string, value string) (string, any) {
					key = strings.ReplaceAll(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".")
					value = strings.ReplaceAll(value, "test", "prod")
					return key, value
				},
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: os.Environ,
			},
		},
		{
			name:  "Environ func is given",
			delim: ".",
			opt: Opt{
				Prefix:      "TEST_",
				EnvironFunc: mockEnviron,
			},
			want: &Env{
				prefix:  "TEST_",
				delim:   ".",
				environ: mockEnviron,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotProvider := Provider(tc.delim, tc.opt)

			// Member variables.
			// Cannot use assert.Equal for the Env itself due to the presense
			// of function member variables.
			assert.Equal(t, tc.want.delim, gotProvider.delim)
			assert.Equal(t, tc.want.prefix, gotProvider.prefix)

			// TransformFunc
			if tc.opt.TransformFunc != nil {
				k, v := gotProvider.transform(tc.key, tc.value)
				assert.Equal(t, tc.expKey, k)
				assert.Equal(t, tc.expValue, v)
			}

			// EnvironFunc
			wantEnv := tc.want.environ()
			gotEnv := gotProvider.environ()
			slices.Sort(wantEnv)
			slices.Sort(gotEnv)
			if !slices.Equal(wantEnv, gotEnv) {
				assert.Fail(t, "Env vars not equal (omitted from message for security)",
					"Want len: %d\nGot len:%d", len(wantEnv), len(gotEnv))
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
				delim:   ".",
				environ: os.Environ,
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
				transform: func(key string, value string) (string, any) {
					return strings.ReplaceAll(strings.ToLower(key), "_", "."), value
				},
				environ: os.Environ,
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
				transform: func(key string, value string) (string, any) {
					return strings.ReplaceAll(strings.ToLower(key), "_", "."), value
				},
				environ: os.Environ,
			},
		},
		{
			name:     "Path value",
			key:      "TEST_DIR",
			value:    "/test/dir/file",
			expKey:   "TEST_DIR",
			expValue: "/test/dir/file",
			env: &Env{
				delim:   ".",
				environ: os.Environ,
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
				transform: func(key string, value string) (string, any) {
					return key, strings.ReplaceAll(strings.ToLower(value), "/", "_")
				},
				environ: os.Environ,
			},
		},
		{
			name:     "Empty value",
			key:      "KEY",
			value:    "",
			expKey:   "KEY",
			expValue: "",
			env: &Env{
				delim:   ".",
				environ: os.Environ,
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
				environ: func() []string {
					return []string{"TEST_OVERRIDE_KEY=TEST_OVERRIDE_VAL"}
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
