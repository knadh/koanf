package koanf_test

import (
	"testing"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

func TestMergeStrategyLoadAppend(t *testing.T) {
	assert := assert.New(t)
	k := koanf.New(delim)

	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"tags":["a"],"nested":{"list":[1]}}`)), json.Parser()))
	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"tags":["b"],"nested":{"list":[2],"x":true}}`)), json.Parser(), koanf.WithMergeAppendSlices()))

	assert.Equal([]any{"a", "b"}, k.Get("tags"))
	assert.Equal([]any{float64(1), float64(2)}, k.Get("nested.list"))
	assert.Equal(true, k.Bool("nested.x"))
}

func TestMergeStrategyLoadByIndex(t *testing.T) {
	assert := assert.New(t)
	k := koanf.New(delim)

	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"providers":[{"id":"google","client_id":1234}]}`)), json.Parser()))
	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"providers":[{"client_secret":"xxxx"},{"id":"okta"}]}`)), json.Parser(), koanf.WithMergeByIndex()))

	slices := k.Slices("providers")
	assert.Len(slices, 2)
	assert.Equal("google", slices[0].String("id"))
	assert.Equal(int64(1234), slices[0].Int64("client_id"))
	assert.Equal("xxxx", slices[0].String("client_secret"))
	assert.Equal("okta", slices[1].String("id"))
}

func TestMergeStrategyLoadByKey(t *testing.T) {
	assert := assert.New(t)
	k := koanf.New(delim)

	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"providers":[{"id":"google","client_id":1234},{"id":"okta","client_id":1}]}`)), json.Parser()))
	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"providers":[{"id":"google","client_secret":"xxxx"},{"id":"azure","client_id":9}]}`)), json.Parser(), koanf.WithMergeByKey("id")))

	slices := k.Slices("providers")
	assert.Len(slices, 3)
	assert.Equal("google", slices[0].String("id"))
	assert.Equal("xxxx", slices[0].String("client_secret"))
	assert.Equal(int64(1234), slices[0].Int64("client_id"))
	assert.Equal("okta", slices[1].String("id"))
	assert.Equal("azure", slices[2].String("id"))
}

func TestMergeStrategyLoadDefaultOverwrite(t *testing.T) {
	assert := assert.New(t)
	k := koanf.New(delim)

	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"tags":["a","b"]}`)), json.Parser()))
	assert.NoError(k.Load(rawbytes.Provider([]byte(`{"tags":["c"]}`)), json.Parser()))
	assert.Equal([]any{"c"}, k.Get("tags"))
}
