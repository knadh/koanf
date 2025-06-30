package env

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockEnviron() []string {
	return []string{"TEST_FOO=bar"}
}

func TestProvider(t *testing.T) {
	type Args struct {
		name     string
		delim    string
		opt      Opt
		key      string
		value    string
		expKey   string
		expValue string
		wantOpt  Opt // Expected options for comparison.
	}
	f := func(args Args) {
		t.Run(args.name, func(t *testing.T) {
			gotProvider := Provider(args.delim, args.opt)
			wantProvider := Provider(args.delim, args.wantOpt)

			// Config variables.
			assert.Equal(t, wantProvider.delim, gotProvider.delim)
			assert.Equal(t, wantProvider.prefix, gotProvider.prefix)

			// Optioonal TransformFunc.
			if args.opt.TransformFunc != nil {
				k, v := gotProvider.transform(args.key, args.value)
				assert.Equal(t, args.expKey, k)
				assert.Equal(t, args.expValue, v)
			}

			// Optional EnvironFunc.
			var (
				wantEnv = wantProvider.environ()
				gotEnv  = gotProvider.environ()
			)
			slices.Sort(wantEnv)
			slices.Sort(gotEnv)
			if !slices.Equal(wantEnv, gotEnv) {
				assert.Fail(t, "Env vars not equal (omitted from message for security)",
					"Want len: %d\nGot len:%d", len(wantEnv), len(gotEnv))
			}
		})
	}

	// Execute the test cases.
	f(Args{name: "No TransformFunc", delim: ".", opt: Opt{Prefix: "TESTVAR_"}, wantOpt: Opt{Prefix: "TESTVAR_"}})
	f(Args{name: "No Opt", delim: "."})
	f(Args{name: "Has EnvironFunc", delim: ".", opt: Opt{Prefix: "TEST_", EnvironFunc: mockEnviron}, wantOpt: Opt{Prefix: "TEST_", EnvironFunc: mockEnviron}})
	f(Args{
		name: "Has TransformFunc", delim: ".", key: "TestKey", value: "TestVal", expKey: "testkey", expValue: "TestVal", opt: Opt{Prefix: "TESTVAR_", TransformFunc: func(k, v string) (string, any) {
			return strings.ToLower(k), v
		}},
		wantOpt: Opt{Prefix: "TESTVAR_"},
	})
	f(Args{
		name: "Has TransformFunc. No Prefix", delim: ".", key: "test_key", value: "test_val", expKey: "TEST.KEY", expValue: "test_val", opt: Opt{
			TransformFunc: func(k, v string) (string, any) {
				return strings.ReplaceAll(strings.ToUpper(k), "_", "."), v
			}},
		wantOpt: Opt{},
	})
	f(Args{
		name: "Has TransformFunc", delim: ".", key: "TEST_KEY", value: "test_val", expKey: "key", expValue: "prod_val", wantOpt: Opt{Prefix: "TEST_"}, opt: Opt{
			Prefix: "TEST_",
			TransformFunc: func(key string, value string) (string, any) {
				key = strings.ReplaceAll(strings.TrimPrefix(strings.ToLower(key), "test_"), "_", ".")
				value = strings.ReplaceAll(value, "test", "prod")
				return key, value
			},
		},
	})

}

func TestRead(t *testing.T) {
	type Args struct {
		name     string
		key      string
		value    string
		expKey   string
		expValue any
		opt      Opt
		delim    string
	}

	f := func(args Args) {
		t.Run(args.name, func(t *testing.T) {
			t.Setenv(args.key, args.value)

			env := Provider(args.delim, args.opt)
			envs, err := env.Read()
			assert.Nil(t, err)

			v, ok := envs[args.expKey]
			assert.True(t, ok)
			assert.Equal(t, args.expValue, v)
		})
	}

	// Execute the test cases.
	f(Args{name: "No TransformFunc", key: "TEST_KEY", value: "TEST_VAL", expKey: "TEST_KEY", expValue: "TEST_VAL", delim: ".", opt: Opt{}})
	f(Args{name: "Path value", key: "TEST_DIR", value: "/test/dir/file", expKey: "TEST_DIR", expValue: "/test/dir/file", delim: "."})
	f(Args{name: "Empty value", key: "KEY", value: "", expKey: "KEY", expValue: "", delim: ".", opt: Opt{}})
	f(Args{
		name: "Has TransformFunc", key: "TEST_KEY", value: "TEST_VAL", expKey: "test.key", expValue: "TEST_VAL", delim: "_", opt: Opt{
			TransformFunc: func(key string, value string) (string, any) {
				return strings.ReplaceAll(strings.ToLower(key), "_", "."), value
			},
		},
	})
	f(Args{
		name: "No TransformFunc. Has Prefix.", key: "TEST_KEY", value: "TEST_VAL", expKey: "test.key", expValue: "TEST_VAL", delim: "/",
		opt: Opt{
			Prefix: "TEST",
			TransformFunc: func(key string, value string) (string, any) {
				return strings.ReplaceAll(strings.ToLower(key), "_", "."), value
			},
		},
	})
	f(Args{
		name: "Transform key", key: "TEST_DIR", value: "/test/dir/file", expKey: "TEST_DIR", expValue: "_test_dir_file", delim: ".",
		opt: Opt{
			TransformFunc: func(key string, value string) (string, any) {
				return key, strings.ReplaceAll(strings.ToLower(value), "/", "_")
			},
		},
	})
	f(Args{
		name: "Has TransformFunc", key: "TEST_KEY", value: "TEST_VAL", expKey: "TEST_OVERRIDE_KEY", expValue: "TEST_OVERRIDE_VAL", delim: ".", opt: Opt{
			EnvironFunc: func() []string {
				return []string{"TEST_OVERRIDE_KEY=TEST_OVERRIDE_VAL"}
			},
		},
	})
	f(Args{
		name: "Has int value", key: "TEST_KEY", value: "123", expKey: "test_key", expValue: 123, delim: ".", opt: Opt{
			EnvironFunc: func() []string {
				return []string{"TEST_KEY=123"}
			},
			TransformFunc: func(key string, value string) (string, any) {
				if key == "TEST_KEY" {
					intval, err := strconv.Atoi(value)
					if err != nil {
						return "", nil
					}
					return "test_key", intval
				}

				return "", ""
			},
		},
	})
}
