package main

import (
	"log"
	"os"
	"strings"

	"github.com/gsingh-ds/koanf"
	"github.com/gsingh-ds/koanf/providers/secretsmanager"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {

	// config for secretsmanager provider
	cofig := secretsmanager.Config{
		SecretId:  "DEV_TEST",
		AWSRegion: os.Getenv("AWS_REGION"),
		AWSAccessKeyID: os.Getenv("AWS_ACCESS_KEY"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		// Type: "map", // provide if type of value is map. it will convert map into: parent: {"child": "value"} -> parent.child = value
	}

	provider := secretsmanager.Provider(cofig , func(s string) string { return strings.ToLower(strings.TrimPrefix(s, "DEV_")) }) 
	
	// Load the provider and parse configuration as JSON.
	if err := k.Load(provider ,nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	k.Print()

	// Watch for all configuration updates.
	provider.Watch(func(event interface{}, err error) {
		if err != nil {
			log.Printf("watch error: %v", err)
			return
		}

		log.Println("config changed. Reloading ...")
		k = koanf.New(".")
		k.Load(provider, nil)
		k.Print()
	})

	log.Println("waiting forever.")
	<-make(chan bool)
}

