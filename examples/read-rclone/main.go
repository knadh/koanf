// Example and test

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rclone"
)

// Global koanf instance. Use . as the key path delimiter. This can be / or anything.
var k = koanf.New(".")

func main() {
	bRemote, err := os.ReadFile("cloud.txt")
	if err != nil {
		log.Fatalf("Cannot read file \"cloud.txt\": %v\n", err)
	}

	remote := string(bRemote)
	remote = strings.TrimSpace(remote)

	f := rclone.Provider(rclone.Config{Remote: remote, File: "mock.json"})
	if err := k.Load(f, json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if strings.Compare("parent1", k.String("parent1.name")) != 0 {
		fmt.Printf("parent1.name: value comparison FAILED\n")
		return
	}

	if k.Int64("parent1.id") != 1234 {
		fmt.Printf("parent1.id: value comparison FAILED\n")
		return
	}

	if strings.Compare("child1", k.String("parent1.child1.name")) != 0 {
		fmt.Printf("parent1.child1.name: value comparison FAILED\n")
		return
	}

	if !k.Bool("parent1.child1.grandchild1.on") {
		fmt.Printf("parent1.child1.grandchild1.on: value comparison FAILED\n")
		return
	}
	
	fmt.Printf("ALL TESTS PASSED\n")
}
