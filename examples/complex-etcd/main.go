// Example and test

package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/etcd/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	clientv3 "go.etcd.io/etcd/client/v3"
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

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 2,
	})

	if err != nil {
		log.Fatal("Cannot create a client by clientv3.New().")
	}
	defer cli.Close()

	keysData := kData.Keys()
	for i := 0; i < len(keysData); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		_, err := cli.Put(ctx, keysData[i], kData.String(keysData[i]))
		cancel()

		if err != nil {
			log.Printf("Couldn't put key.")
		}
	}

	// Single key/value.
	var (
		sKey = "single_key"
		sVal = "single_val"
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	_, err = cli.Put(ctx, sKey, sVal)
	defer cancel()

	if err != nil {
		log.Printf("Couldn't put key.")
	}

	providerCfg := etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
		Prefix:      false,
		Key:         "single_key",
	}
	provider, err := etcd.Provider(providerCfg)
	if err != nil {
		log.Fatalf("Failed to instantiate etcd provider: %v", err)
	}

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

	providerCfg = etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
		Prefix:      true,
		Key:         "parent",
	}
	provider, err = etcd.Provider(providerCfg)
	if err != nil {
		log.Fatalf("Failed to instantiate etcd provider: %v", err)
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
	// etcdctl get --prefix child

	if err := kReq.Load(file.Provider("req2.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	providerCfg = etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
		Prefix:      true,
		Key:         "child",
	}

	provider, err = etcd.Provider(providerCfg)
	if err != nil {
		log.Fatalf("Failed to instantiate etcd provider: %v", err)
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
			fmt.Printf("Second request: key comparison FAILED\n")
			return
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("Second request: value comparison FAILED\n")
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

	providerCfg = etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
		Prefix:      true,
		Limit:       true,
		NLimit:      4,
		Key:         "child",
	}

	provider, err = etcd.Provider(providerCfg)
	if err != nil {
		log.Fatalf("Failed to instantiate etcd provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	keysReq = kReq.Keys()
	keysCheck = kCheck.Keys()

	if len(keysReq) != len(keysCheck) {
		fmt.Printf("Third request: keys FAILED\n")
		return
	}

	for i := 0; i < len(keysReq); i++ {
		if strings.Compare(keysReq[i], keysCheck[i]) != 0 {
			fmt.Printf("Third request: key comparison FAILED\n")
			return
		}

		if strings.Compare(kReq.String(keysReq[i]), kCheck.String(keysCheck[i])) != 0 {
			fmt.Printf("Third request: value comparison FAILED\n")
			return
		}
	}

	fmt.Printf("Third (combined) request test passed.\n")

	kCheck.Delete("")

	// Watch test

	sKey = "child"
	providerCfg = etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
		Prefix:      true,
		Key:         "child",
	}

	provider, err = etcd.Provider(providerCfg)
	if err != nil {
		log.Fatalf("Failed to instantiate etcd provider: %v", err)
	}

	if err := kCheck.Load(provider, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	changedC := make(chan string, 1)

	provider.Watch(func(event any, err error) {
		if err != nil {
			fmt.Printf("Unexpected error: %v", err)
			return
		}

		kCheck.Load(provider, nil)
		changedC <- kCheck.String(string(event.(*clientv3.Event).Kv.Key))
	})

	var newVal string = "Brian"
	cmd := exec.Command("etcdctl", "put", "child1", newVal)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Compare(newVal, <-changedC) != 0 {
		fmt.Printf("Watch failed: new value comparison FAILED\n")
		return
	}

	newVal = "Kate"
	cmd = exec.Command("etcdctl", "put", "child2", newVal)
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
