package env

import (
	"github.com/stretchr/testify/assert"
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
