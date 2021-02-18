package main

import (
	"fmt"
	"log"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/awssm"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	s, err := awssm.Provider(awssm.Config{
		SecretName: os.Getenv("AWS_REGION"),
		Region:     os.Getenv("AWS_SECRET_NAME"),
	})
	if err != nil {
		log.Fatalf("error loading AwsSecrets Manager Provider: %v", err)
	}
	if err := k.Load(s, json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("parent's name is = ", k.String("parent1.name"))
	fmt.Println("parent's ID is = ", k.Int("parent1.id"))
}
