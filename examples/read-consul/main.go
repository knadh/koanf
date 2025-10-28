package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/consul/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	kData  = koanf.New(".")
	kReq   = koanf.New(".")
	kCheck = koanf.New(".")
)

func main() {
	if err := kData.Load(file.Provider("data.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	kv := cli.KV()

	keysData := kData.Keys()
	for _, key := range keysData {
		newPair := &api.KVPair{Key: key, Value: []byte(kData.String(key))}
		_, err = kv.Put(newPair, nil)

		if err != nil {
			log.Printf("Couldn't put key.")
		}
	}

	// Single key/value.
	var (
		sKey string = "single_key"
		sVal string = "single_val"
	)

	newPair := &api.KVPair{Key: sKey, Value: []byte(sVal)}
	_, err = kv.Put(newPair, nil)

	if err != nil {
		log.Printf("Couldn't put key.")
	}

	provider, err := consul.Provider(consul.Config{
		Key:      sKey,
		Recurse:  false,
		Detailed: false,
		Cfg:      api.DefaultConfig(),
	})
	if err != nil {
		log.Fatalf("Failed to instantiate consul provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if len(kCheck.Keys()) != 1 {
		fmt.Printf("Single key: FAILED\n")
		return
	}

	if strings.Compare(sKey, kCheck.Keys()[0]) != 0 {
		fmt.Printf("Single key: key comparison FAILED\n")
		return
	}

	if strings.Compare(sVal, kCheck.String(kCheck.Keys()[0])) != 0 {
		fmt.Printf("Single key: value comparison FAILED\n")
		return
	}

	fmt.Printf("\nSingle key test passed.\n")
	kCheck.Delete("")

	// first request test
	// analog of the command:
	// consul kv get -recurse parent

	if err := kReq.Load(file.Provider("req1.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	provider, err = consul.Provider(consul.Config{
		Key:      "parent",
		Recurse:  true,
		Detailed: false,
		Cfg:      api.DefaultConfig(),
	})
	if err != nil {
		log.Fatalf("Failed to instantiate consul provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keysReq := kReq.Keys()
	keysCheck := kCheck.Keys()

	if len(keysReq) != len(keysCheck) {
		fmt.Printf("First request: keys FAILED\n")
		return
	}

	for i := 0; i < len(keysReq); i++ {
		if strings.Compare(keysReq[i], keysCheck[i]) != 0 {
			fmt.Printf("First request: key comparison FAILED\n")
			return
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("First request: value comparison FAILED\n")
			return
		}
	}

	fmt.Printf("First request test passed.\n")
	kReq.Delete("")
	kCheck.Delete("")

	// second request test
	// analog of the command:
	// consul kv get -recurse child

	if err := kReq.Load(file.Provider("req2.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	provider, err = consul.Provider(consul.Config{
		Key:      "child",
		Recurse:  true,
		Detailed: false,
		Cfg:      api.DefaultConfig(),
	})
	if err != nil {
		log.Fatalf("Failed to instantiate consul provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keysReq = kReq.Keys()
	keysCheck = kCheck.Keys()

	if len(keysReq) != len(keysCheck) {
		fmt.Printf("Second request: keys FAILED\n")
		return
	}

	for i := 0; i < len(keysReq); i++ {
		if strings.Compare(keysReq[i], keysCheck[i]) != 0 {
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("Second request: value comparison FAILED\n")
			return
		}
	}

	fmt.Printf("Second request test passed.\n")
	kCheck.Delete("")

	// adding metadata: age (flags)
	newPair = &api.KVPair{Key: "parent1", Flags: uint64(42), Value: []byte("father")}
	_, err = kv.Put(newPair, nil)
	if err != nil {
		log.Printf("Couldn't put key with flags.")
	}

	// Single key detailed test.
	// analog of the command:
	// consul kv get -detailed parent1

	sKey = "parent1"
	sFlags := uint64(42)
	sVal = "father"

	provider, err = consul.Provider(consul.Config{
		Key:      sKey,
		Recurse:  false,
		Detailed: true,
		Cfg:      api.DefaultConfig(),
	})
	if err != nil {
		log.Fatalf("Failed to instantiate consul provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		fmt.Printf("error loading config: %v", err)
		return
	}

	if sFlags != uint64(kCheck.Int64("parent1.Flags")) {
		fmt.Printf("Single detailed key: flags (metadata: age) comparison FAILED\n")
		return
	}

	if strings.Compare(sVal, kCheck.String("parent1.Value")) != 0 {
		fmt.Printf("Single detailed key: value comparison FAILED\n")
		return
	}

	fmt.Printf("\nDetailed single key test passed.\n")

	kCheck.Delete("")

	// Detailed request (recurse) test.
	// analog of the command:
	// consul kv get -detailed -recurse parent

	sKey = "parent"

	provider, err = consul.Provider(consul.Config{
		Key:      sKey,
		Recurse:  true,
		Detailed: true,
		Cfg:      api.DefaultConfig(),
	})
	if err != nil {
		log.Fatalf("Failed to instantiate consul provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		fmt.Printf("error loading config: %v", err)
		return
	}

	if sFlags != uint64(kCheck.Int64("parent1.Flags")) {
		fmt.Printf("Single detailed key: flags (metadata: age) comparison FAILED\n")
		return
	}

	if strings.Compare(sVal, kCheck.String("parent1.Value")) != 0 {
		fmt.Printf("Single key: value comparison FAILED\n")
		return
	}

	sFlags = uint64(0)
	sVal = "mother"

	if sFlags != uint64(kCheck.Int64("parent2.Flags")) {
		fmt.Printf("Single detailed key: flags (metadata: age) comparison FAILED\n")
		return
	}

	if strings.Compare(sVal, kCheck.String("parent2.Value")) != 0 {
		fmt.Printf("Single key: value comparison FAILED\n")
		return
	}

	fmt.Printf("\nDetailed request (recurse) test passed.\n")

	kCheck.Delete("")

	// Watch test

	sKey = "parent"
	provider, err = consul.Provider(consul.Config{
		Key:      sKey,
		Recurse:  true,
		Detailed: false,
		Cfg:      api.DefaultConfig(),
	})
	if err != nil {
		log.Fatalf("Failed to instantiate consul provider: %v", err)
	}

	// Getting the old value
	kCheck.Load(provider, nil)
	oldVal := kCheck.String("parent1")

	changedC := make(chan string, 1)

	provider.Watch(func(event any, err error) {
		if err != nil {
			fmt.Printf("Unexpected error: %v", err)
			return
		}

		kCheck.Load(provider, nil)
		// skip the first call
		if strings.Compare(oldVal, kCheck.String("parent1")) != 0 {
			changedC <- kCheck.String("parent1")
		}
	})

	// changing
	var newVal string = "dad"
	cmd := exec.Command("consul", "kv", "put", "parent1", newVal)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Compare(newVal, <-changedC) != 0 {
		fmt.Printf("Watch failed: new value comparison FAILED\n")
		return
	}
	fmt.Printf("Watch test passed.\n")

	fmt.Printf("ALL TESTS PASSED\n")
}
