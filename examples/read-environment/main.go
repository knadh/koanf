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
	// For example, env vars: MYVAR_TYPE and MYVAR_PARENT1_CHILD1_NAME
	// will be merged into the "type" and the nested "parent1.child1.name"
	// keys in the config file here as we lowercase the key,
	// replace `_` with `.` and strip the MYVAR_ prefix so that
	// only "parent1.child1.name" remains.
	k.Load(env.Provider("MYVAR_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "MYVAR_")), "_", ".", -1)
	}), nil)

	// Use ProviderWithValue() to process both keys and values into types other than strings,
	// for example, turn space separated env vars into slices.
	// k.Load(env.ProviderWithValue("MYVAR_", ".", func(s string, v string) (string, interface{}) {
	// 	// Strip out the MYVAR_ prefix and lowercase and get the key while also replacing
	// 	// the _ character with . in the key (koanf delimeter).
	// 	key := strings.Replace(strings.ToLower(strings.TrimPrefix(s, "MYVAR_")), "_", ".", -1)

	// 	// If there is a space in the value, split the value into a slice by the space.
	// 	if strings.Contains(v, " ") {
	// 		return key, strings.Split(v, " ")
	// 	}

	// 	// Otherwise, return the plain string.
	// 	return key, v
	// }), nil)

	fmt.Println("name is = ", k.String("parent1.child1.name"))
}
