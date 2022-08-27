package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/properties"
	"github.com/knadh/koanf/providers/file"
)

var k = koanf.New(".")

func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("file1.properties"), properties.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Printf("website: %s\n", k.String("website"))
}
