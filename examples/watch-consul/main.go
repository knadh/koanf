package main

import (
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

	p.Watch(func(event interface{}, err error) {
		if err != nil {
			log.Printf("watch error: %v", err)
			return
		}

		log.Println("KV set changed. Updating...")
		k = koanf.New(".")
		k.Load(p, nil)
		k.Print()
	})

	log.Println("waiting forever. Try to change \"parent\" values.")
	<-make(chan bool)
}
