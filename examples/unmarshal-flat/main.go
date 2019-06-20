package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
)

// Global koanf instance. Use . as the key path delimiter. This can be / or anything.
var k = koanf.New(".")

func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("mock/mock.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	type rootFlat struct {
		Type                        string            `koanf:"type"`
		Empty                       map[string]string `koanf:"empty"`
		Parent1Name                 string            `koanf:"parent1.name"`
		Parent1ID                   int               `koanf:"parent1.id"`
		Parent1Child1Name           string            `koanf:"parent1.child1.name"`
		Parent1Child1Type           string            `koanf:"parent1.child1.type"`
		Parent1Child1Empty          map[string]string `koanf:"parent1.child1.empty"`
		Parent1Child1Grandchild1IDs []int             `koanf:"parent1.child1.grandchild1.ids"`
		Parent1Child1Grandchild1On  bool              `koanf:"parent1.child1.grandchild1.on"`
	}

	// Unmarshal the whole root with FlatPaths: True.
	var o1 rootFlat
	k.UnmarshalWithConf("", &o1, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true})
	fmt.Println(o1)

	// Unmarshal a child structure of "parent1".
	type subFlat struct {
		Name                 string            `koanf:"name"`
		ID                   int               `koanf:"id"`
		Child1Name           string            `koanf:"child1.name"`
		Child1Type           string            `koanf:"child1.type"`
		Child1Empty          map[string]string `koanf:"child1.empty"`
		Child1Grandchild1IDs []int             `koanf:"child1.grandchild1.ids"`
		Child1Grandchild1On  bool              `koanf:"child1.grandchild1.on"`
	}

	var o2 subFlat
	k.UnmarshalWithConf("parent1", &o2, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true})
	fmt.Println(o2)
}
