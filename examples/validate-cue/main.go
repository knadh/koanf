package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/validators/vcue"
)

var k = koanf.New(".")

func main() {
	f := file.Provider("mock/kub.yml")
	if err := k.Load(f, yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Printf("API version: %s\n", k.String("apiVersion"))

	v := vcue.Validator("scheme.cue", vcue.BlockAll)

	err := k.Validate(v)
	if err == nil {
		fmt.Printf("Data correct.\n")
	}
}
