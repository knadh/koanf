package etcd

import (
	"context"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"testing"
	"time"
)

func TestEtcd_ReadBytes(t *testing.T) {
	// Start an embedded etcd server
	cfg := embed.NewConfig()
	cfg.Logger = "zap"
	cfg.LogLevel = "error"
	cfg.Dir = "/tmp/test-etcd"
	cfg.WalDir = ""
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	// Create a client connection to the embedded server
	endpoints := []string{e.Clients[0].Addr().String()}

	// Use the client to put key-value pairs in the etcd server
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "Read non-empty data",
			key:   "test-key",
			value: "---\ntype: yml\nparent1:\n  name: parent1\n  id: 1234\n  child1:\n    name: child1\n    type: yml\n    grandchild1:\n      ids:\n        - 1\n        - 2\n        - 3\n      \"on\": true\n    empty: {}\n  strmap:\n    key1: val1\n    key2: val2\n    key3: val3\n  strsmap:\n    key1:\n      - val1\n      - val2\n      - val3\n    key2:\n      - val4\n      - val5\n  intmap:\n    key1: 1\n    key2: 1\n    key3: 1\n  floatmap:\n    key1: 1.1\n    key2: 1.2\n    key3: 1.3\n  boolmap:\n    ok1: true\n    ok2: true\n    notok3: false\nparent2:\n  name: parent2\n  id: 5678\n  child2:\n    name: child2\n    grandchild2:\n      ids:\n        - 4\n        - 5\n        - 6\n      \"on\": true\n    empty: {}\norphan:\n  - red\n  - blue\n  - orange\nempty: {}\nbools:\n  - true\n  - false\n  - true\nintbools:\n  - 1\n  - 0\n  - 1\nstrbools:\n  - \"1\"\n  - t\n  - f\nstrbool: \"1\"\ntime: \"2019-01-01\"\nduration: \"3s\"\nnegative_int: -1234\n",
		},
		{
			name:  "Read empty data",
			key:   "test-key-empty",
			value: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := clientv3.New(clientv3.Config{
				Endpoints: endpoints,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer cli.Close()

			_, err = cli.Put(context.Background(), tt.key, tt.value)
			if err != nil {
				t.Fatal(err)
			}

			provider := Provider(Config{
				Endpoints:   endpoints,
				DialTimeout: 5 * time.Second,
				Key:         tt.key,
			})

			var k = koanf.New(".")
			if err := k.Load(provider, yaml.Parser()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
