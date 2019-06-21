package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

// Global koanf instance. Use . as the key path delimiter. This can be / or anything.
var k = koanf.New(".")

func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("mock/mock.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Load environment variables and merge into the loaded config.
	// "MYVAR" is the prefix to filter the env vars by.
	// "." is the delimiter used to represent the key hierarchy in env vars.
	// The (optional, or can be nil) function can be used to transform
	// the env var names, for instance, to lowercase them.
	//
	// For example, env vars: MYVAR_TYPE and MYVAR_PARENT1.CHILD1.NAME
	// will me merged into the "type" and the nested "parent1.child1.name"
	// keys in the config file.
	k.Load(env.Provider("MYVAR_", ".", func(s string) string {
		return strings.ToLower(s)
	}), nil)

	fmt.Println("name is =", k.String("parent1.child1.name"))
}
