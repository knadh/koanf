package main

import (
	"log"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/appconfig"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	provider := appconfig.Provider(appconfig.Config{
		Application:   os.Getenv("AWS_APPCONFIG_APPLICATION"),
		ClientID:      os.Getenv("AWS_APPCONFIG_CLIENT_ID"),
		Configuration: os.Getenv("AWS_APPCONFIG_CONFIG_NAME"),
		Environment:   os.Getenv("AWS_APPCONFIG_ENVIRONMENT"),
	})
	// Load the provider and parse configuration as JSON.
	if err := k.Load(provider, json.Parser()); err != nil {
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
		k.Load(provider, json.Parser())
		k.Print()
	})

	log.Println("waiting forever.")
	<-make(chan bool)
}
