package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/consul"
	"github.com/hashicorp/consul/api"
)

var k = koanf.New(".")

func main() {
	p := consul.Provider(consul.Config{
		Key: "parent",
		Recurse: true,
		Detailed: false,
		CConfig: api.DefaultConfig(),
	})

	if err := k.Load(p, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("parent's name is = ", k.String("parent1"))

	// Watch the file and get a callback on change. The callback can do whatever,
	// like re-load the configuration.
	// File provider always returns a nil `event`.
	p.Watch(func(event interface{}, err error) {
		if err != nil {
			log.Printf("watch error: %v", err)
			return
		}

		// Throw away the old config and load a fresh copy.
		log.Println("config changed. Reloading ...")
		k = koanf.New(".")
		k.Load(p, nil)
		fmt.Printf("parent1: %s\n", k.String("parent1"))
		fmt.Printf("parent2: %s\n", k.String("parent2"))
	})

	// Block forever (and manually make a change to mock/mock.json) to
	// reload the config.
	log.Println("waiting forever. Try making a change to mock/mock.json to live reload")
	<-make(chan bool)
}
