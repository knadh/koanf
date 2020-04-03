package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
)

// Global koanf instance. Use . as the key path delimiter. This can be / or anything.
var (
	k      = koanf.New(".")
	parser = json.Parser()
)

func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("mock/mock.json"), parser); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Structure to unmarshal nested conf to.
	type childStruct struct {
		Name       string            `koanf:"name"`
		Type       string            `koanf:"type"`
		Empty      map[string]string `koanf:"empty"`
		GrandChild struct {
			Ids []int `koanf:"ids"`
			On  bool  `koanf:"on"`
		} `koanf:"grandchild1"`
	}

	var out childStruct

	// Quick unmarshal.
	k.Unmarshal("parent1.child1", &out)
	fmt.Println(out)

	// Unmarshal with advanced config.
	out = childStruct{}
	k.UnmarshalWithConf("parent1.child1", &out, koanf.UnmarshalConf{Tag: "koanf"})
	fmt.Println(out)

	// Marshal the instance back to JSON.
	// The paser instance can be anything, eg: json.Paser(),
	// yaml.Parser() etc.
	b, _ := k.Marshal(parser)
	fmt.Println(string(b))
}
