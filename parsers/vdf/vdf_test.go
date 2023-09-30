package vdf

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/andygrunwald/vdf"
	"github.com/stretchr/testify/require"
)

func TestVDF_Marshal(t *testing.T) {
	type testCase struct {
		name   string
		input  map[string]interface{}
		output []byte
	}
	tt := testCase{name: "Valid VDF",
		input: map[string]interface{}{"map1": map[string]interface{}{"key1": "value1", "key2": "value2",
			"map2": map[string]interface{}{"key3": "value3", "key4": "value4"}},
			"map3": map[string]interface{}{"key5": "value5", "key6": "value6"}},
		output: []byte(`
		"map3"
		{
			"key5"		"value5"
			"key6"		"value6"
		}
		"map1"
		{
			"key1"		"value1"
			"key2"		"value2"
			"map2"
			{
				"key3"		"value3"
				"key4"		"value4"
			}
		}
		`),
	}

	t.Run(tt.name, func(t *testing.T) {
		d, err := newDumper(tt.input)
		require.NoError(t, err)
		reflect.DeepEqual(d, tt.output)
	})
}
func TestVDF_Unmarshal(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
		want  func(got map[string]interface{}, err error)
	}{
		{
			name:  "Empty VDF",
			input: []byte(""),
			want: func(got map[string]interface{}, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "Valid VDF",
			input: []byte(`"SaveFile"
							{
								"team1"		"ciccio"

								"team2"		"pasticcio"
							}
						`),
			want: func(got map[string]interface{}, err error) {
				require.NoError(t, err)
				saveFile := got["SaveFile"].(map[string]interface{})
				require.Equal(t, "ciccio", saveFile["team1"])
				require.Equal(t, "pasticcio", saveFile["team2"])
			},
		},
		{
			name: "Corrupted VDF",
			input: []byte(`"SaveFile"
							{
								"team1"		"ciccio"
								"team2"		"pasticcio	
						`),
			want: func(got map[string]interface{}, err error) {
				require.True(t, errors.Is(err, vdf.ErrNotValidFormat))
			},
		},
		{
			name: "Corrupted VDF no curly brace",
			input: []byte(`"SaveFile"
						
							"team1"		"ciccio"
							"team2"		"pasticcio	
						`),
			want: func(got map[string]interface{}, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "Valid VDF broken comment",
			input: []byte(`"Broken Comment"
							{
							// This is a valid comment. uri and timeout will be parsed as expected
							"uri" "http://127.0.0.1:3456"
							"timeout" "8.0"
							/ This is a broken/invalid comment. Parsing will stop here and buffer ignored
							"buffer"  "0.3"
						}`),
			want: func(got map[string]interface{}, err error) {
				require.NoError(t, err)
				expected := map[string]interface{}{
					"Broken Comment": map[string]interface{}{
						"uri":     "http://127.0.0.1:3456",
						"timeout": "8.0",
					},
				}
				require.Equal(t, expected, got)
			},
		},
	}
	getReader := func(input []byte) (io.Reader, error) {
		return bytes.NewReader(input), nil
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := getReader(tt.input)
			require.NoError(t, err)
			p := vdf.NewParser(reader)
			got, err := p.Parse()
			tt.want(got, err)
		})
	}

}
