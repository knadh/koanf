package main

import (
	"log"
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/consul"
)

var k = koanf.New(".")

func main() {
	provider := consul.Provider(consul.Config {
		Key: "parent1",
		Recurse: false,
		Detailed: false,
	})

	if err := k.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Printf("parent1: %s\n", k.String("parent1"))
}
