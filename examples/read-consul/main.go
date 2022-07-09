package main

import (
	"log"
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/consul"
	"github.com/hashicorp/consul/api"
)

var k = koanf.New(".")

func main() {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	kv := cli.KV()

	newPair := &api.KVPair{Key: "parent1", Value: []byte("father")}
	_, err = kv.Put(newPair, nil)
	if err != nil {
		panic(err)
	}

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
