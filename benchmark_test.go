package koanf_test

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"io/ioutil"
	"testing"
)

var jsonParser = json.Parser()
var mockLargeDataProvider koanf.Provider
var largeDataInstance *koanf.Koanf

func init() {
	mockLargeData, err := ioutil.ReadFile("mock/mock-flat.json")
	if err != nil {
		panic(err)
	}

	mockLargeDataProvider = rawbytes.Provider(mockLargeData)

	largeDataInstance = koanf.New(delim)
	if err := largeDataInstance.Load(mockLargeDataProvider, jsonParser); err != nil {
		panic(err)
	}
}

func BenchmarkInstantiation(b *testing.B) {
	koanfs := make([]*koanf.Koanf, b.N)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		k := koanf.New(delim)
		if err := k.Load(mockLargeDataProvider, jsonParser); err != nil {
			b.Fatalf("Unexpected error: %+v", k)
		}
		koanfs[n] = k
	}

	b.Logf("%d", len(koanfs))
}

func BenchmarkGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if largeDataInstance.Int("serve.admin.socket.mode") != 493 {
			b.Fatalf("Expected 493")
		}
	}
}
