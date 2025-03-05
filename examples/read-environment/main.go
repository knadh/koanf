package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

// Run this example:
//
//	go run ./examples/read-environment
//
// Try setting environment variables to see what changes (and what doesn't!):
//
//	MYVAR_PARENT1_CHILD1_NAME=FooBar go run ./examples/read-environment
//	MYVAR_TIME=2020-02-02 go run ./examples/read-environment
//	MYVAR_PARENT1_CHILD1_GRANDCHILD1_IDS="3 2 1" go run ./examples/read-environment
func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("mock/mock.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Load environment variables and merge into the loaded config.
	// "MYVAR" is the prefix to filter the env vars by.
	// "." is the delimiter used to represent the key hierarchy in env vars.
	// The optional TransformFunc can be used to transform the env var names,
	// for instance, to lowercase them. Values can also be modified into types
	// other than strings, for example, to turn space separated env vars into
	// slices.
	//
	// will be merged into the "type" and the nested "parent1.child1.name"
	// For example, env vars: MYVAR_TYPE and MYVAR_PARENT1_CHILD1_NAME
	// keys in the config file here as we lowercase the key,
	// replace `_` with `.` and strip the MYVAR_ prefix so that
	// only "parent1.child1.name" remains.
	//
	// The optional EnvironFunc can be used to control what environment
	// variables are read by Koanf. In this example, we ensure that the time
	// variable cannot be overridden by env vars.
	k.Load(env.Provider(".", env.Opt{
		Prefix: "MYVAR_",
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(
				strings.TrimPrefix(k, "MYVAR_")), "_", ".")

			// If there is a space in the value, split the value into a slice by the space.
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}
			return k, v
		},
		EnvironFunc: func() []string {
			return slices.DeleteFunc(os.Environ(), func(s string) bool {
				return strings.HasPrefix(s, "MYVAR_TIME")
			})
		},
	}), nil)

	fmt.Println("name is =", k.String("parent1.child1.name"))
	fmt.Println("time is =", k.Time("time", time.DateOnly))
	fmt.Println("ids are =", k.Strings("parent1.child1.grandchild1.ids"))
}
