// Example and test

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
	kData = koanf.New(".")
	kReq = koanf.New(".")
	kCheck = koanf.New(".")
)

func main() {

	if err := kData.Load(file.Provider("data.json"), json.Parser()); err != nil {
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

	keysData := kData.Keys();

	for i := 0; i < len(keysData); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 2)
		_, err := cli.Put(ctx, keysData[i], kData.String(keysData[i]))
		cancel()

		if err != nil {
			log.Printf("Couldn't put key.")
		}
	}

	// single key/value

	var sKey = "single_key"
	var sVal = "single_val"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 2)
	_, err = cli.Put(ctx, sKey, sVal)
	cancel()

	if err != nil {
		log.Printf("Couldn't put key.")
	}

	providerCfg := etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: false,
		Keypath: "single_key",
	}

	provider := etcd.Provider(providerCfg)

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if len(kCheck.Keys()) != 1 {
		fmt.Println(len(kCheck.Keys()))
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
	// etcdctl get --prefix parent

	if err := kReq.Load(file.Provider("req1.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	providerCfg = etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: true,
		Keypath: "parent",
	}

	provider = etcd.Provider(providerCfg)

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keysReq := kReq.Keys();
	keysCheck := kCheck.Keys();

	if len(keysReq) != len(keysCheck) {
		fmt.Printf("FAILED\n")
		return
	}

	for i := 0; i < len(keysReq); i++ {
		if strings.Compare(keysReq[i], keysCheck[i]) != 0 {
			fmt.Printf("FAILED\n")
			return
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("FAILED\n")
			return
		}
	}

	fmt.Printf("First request test passed.\n")
	kReq.Delete("")
	kCheck.Delete("")

	// second request test
	// analog of the command:
	// etcdctl get --prefix child

	if err := kReq.Load(file.Provider("req2.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	providerCfg = etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: true,
		Keypath: "child",
	}

	provider = etcd.Provider(providerCfg)

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keysReq = kReq.Keys();
	keysCheck = kCheck.Keys();

	if len(keysReq) != len(keysCheck) {
		fmt.Printf("FAILED\n")
		return
	}

	for i := 0; i < len(keysReq); i++ {
		if strings.Compare(keysReq[i], keysCheck[i]) != 0 {
			fmt.Printf("FAILED\n")
			return
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("FAILED\n")
			return
		}
	}

	fmt.Printf("Second request test passed.\n")
	kReq.Delete("")
	kCheck.Delete("")

	// third (combined prefix + limit) request test
	// analog of the command:
	// etcdctl get --prefix child --limit=4

	if err := kReq.Load(file.Provider("req3.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	providerCfg = etcd.Config {
		Endpoints: []string { "localhost:2379" },
		DialTimeout: time.Second * 5,
		Prefix: true,
		Limit: true,
		nLimit: 4,
		Keypath: "child",
	}

	provider = etcd.Provider(providerCfg)

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keysReq = kReq.Keys();
	keysCheck = kCheck.Keys();

	if len(keysReq) != len(keysCheck) {
		fmt.Printf("FAILED\n")
		return
	}

	for i := 0; i < len(keysReq); i++ {
		if strings.Compare(keysReq[i], keysCheck[i]) != 0 {
			fmt.Printf("FAILED\n")
			return
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("FAILED\n")
			return
		}
	}

	fmt.Printf("Third (combined) request test passed.\n")

	fmt.Printf("ALL TESTS PASSED\n")
}
