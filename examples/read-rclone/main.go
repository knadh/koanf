package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rclone"
)

// Global koanf instance. Use . as the key path delimiter. This can be / or anything.
var k = koanf.New(".")

func main() {
	f := rclone.Provider(rclone.Config{Remote: "godrive1", File: "mock.json"})
	if err := k.Load(f, json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	f2 := rclone.Provider(rclone.Config{Remote: "minio1", File: "bucket1/mock.yml"})
	if err := k.Load(f2, yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("parent's name is = ", k.String("parent1.name"))
	fmt.Println("parent's ID is = ", k.Int("parent1.id"))
	fmt.Println("object name: ", k.String("object1.name"))
	fmt.Println("object embedded name: ", k.String("object1.embedded1.name"))
	fmt.Println("object nest: ", k.Bool("object1.embedded1.nest1.on"))
}
