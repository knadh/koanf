package main

import (
	"time"
	"log"
	"fmt"
	"context"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	k_data = koanf.New(".")
	k_req = koanf.New(".")
	k_check = koanf.New(".")
)

func main() {

	if err := k_data.Load(file.Provider("data.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	cli, err := clientv3.New(clientv3.Config {
		Endpoints:		[]string{ "localhost:2379" },
		DialTimeout:	time.Second * 2,
	})

	if err != nil {
		log.Fatal("Cannot create a client by clientv3.New().")
	}
	defer cli.Close()

	keys_data := k_data.Keys();

	for i := 0; i < len(keys_data); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 2)
		_, err := cli.Put(ctx, keys_data[i], k_data.String(keys_data[i]))
		cancel()

		if err != nil {
			log.Printf("Couldn't put key.")
		}
	}

	// single key/value

	var s_key = "single_key"
	var s_val = "single_val"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 2)
	_, err = cli.Put(ctx, s_key, s_val)
	cancel()

	if err != nil {
		log.Printf("Couldn't put key.")
	}

	provider_cfg := etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: false,
		Keypath: "single_key",
	}

	provider := etcd.Provider(provider_cfg)

	if err := k_check.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if len(k_check.Keys()) != 1 {
		fmt.Println(len(k_check.Keys()))
		fmt.Printf("Single key: FAILED\n")
		return
	}

	if strings.Compare(s_key, k_check.Keys()[0]) != 0 {
		fmt.Printf("Single key: key comparison FAILED\n")
		return
	}

	if strings.Compare(s_val, k_check.String(k_check.Keys()[0])) != 0 {
		fmt.Printf("Single key: value comparison FAILED\n")
		return
	}

	fmt.Printf("\nSingle key test passed.\n")
	k_check.Delete("")

	// first request test

	if err := k_req.Load(file.Provider("req1.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	provider_cfg = etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: true,
		Keypath: "parent",
	}

	provider = etcd.Provider(provider_cfg)

	if err := k_check.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keys_req := k_req.Keys();
	keys_check := k_check.Keys();

	if len(keys_req) != len(keys_check) {
		fmt.Printf("FAILED\n")
		return
	}

	for i := 0; i < len(keys_req); i++ {
		if strings.Compare(keys_req[i], keys_check[i]) != 0 {
			fmt.Printf("FAILED\n")
			return
		}

		if strings.Compare(k_req.String(keys_req[i]), k_check.String(keys_check[i])) != 0 {
			fmt.Printf("FAILED\n")
			return
		}
	}

	fmt.Printf("First request test passed.\n")
	k_req.Delete("")
	k_check.Delete("")

	// second request test

	if err := k_req.Load(file.Provider("req2.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	provider_cfg = etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: true,
		Keypath: "child",
	}

	provider = etcd.Provider(provider_cfg)

	if err := k_check.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keys_req = k_req.Keys();
	keys_check = k_check.Keys();

	if len(keys_req) != len(keys_check) {
		fmt.Printf("FAILED\n")
		return
	}

	for i := 0; i < len(keys_req); i++ {
		if strings.Compare(keys_req[i], keys_check[i]) != 0 {
			fmt.Printf("FAILED\n")
			return
		}

		if strings.Compare(k_req.String(keys_req[i]), k_check.String(keys_check[i])) != 0 {
			fmt.Printf("FAILED\n")
			return
		}
	}

	fmt.Printf("Second request test passed.\n")
	k_req.Delete("")
	k_check.Delete("")

	// third (combined prefix + limit) request test

	if err := k_req.Load(file.Provider("req3.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	provider_cfg = etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: true,
		Limit: true,
		NLimit: 4,
		Keypath: "child",
	}

	provider = etcd.Provider(provider_cfg)

	if err := k_check.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keys_req = k_req.Keys();
	keys_check = k_check.Keys();

	if len(keys_req) != len(keys_check) {
		fmt.Printf("FAILED\n")
		return
	}

	for i := 0; i < len(keys_req); i++ {
		if strings.Compare(keys_req[i], keys_check[i]) != 0 {
			fmt.Printf("FAILED\n")
			return
		}

		if strings.Compare(k_req.String(keys_req[i]), k_check.String(keys_check[i])) != 0 {
			fmt.Printf("FAILED\n")
			return
		}
	}

	fmt.Printf("Third (combined) request test passed.\n")

	fmt.Printf("ALL TESTS PASSED\n")
}
