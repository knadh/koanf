package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf/parsers/huml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	// Load HUML config.
	if err := k.Load(file.Provider("../../mock/mock.huml"), huml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("app name = ", k.String("app_name"))
	fmt.Println("version = ", k.String("version"))
	fmt.Println("debug = ", k.Bool("debug"))
	fmt.Println("port = ", k.Int("port"))
	fmt.Println("timeout = ", k.Float64("timeout"))
	
	// Access arrays
	fmt.Println("tags = ", k.Strings("tags"))
	fmt.Println("ports = ", k.Ints("ports"))
	
	// Access nested objects
	fmt.Println("database host = ", k.String("database.host"))
	fmt.Println("database port = ", k.Int("database.port"))
	fmt.Println("database ssl = ", k.Bool("database.ssl"))
	
	// Access deeply nested values
	fmt.Println("http port = ", k.Int("server.http.port"))
	fmt.Println("https cert = ", k.String("server.https.cert"))
	
	// Access environment-specific settings
	fmt.Println("dev debug = ", k.Bool("environments.development.debug"))
	fmt.Println("prod log level = ", k.String("environments.production.log_level"))
	
	// Print the entire config
	fmt.Println("\nFull configuration:")
	k.Print()
}