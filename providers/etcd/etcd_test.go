package etcd

import (
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
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

	provider, err := Provider(Config{
		ClientConfig: clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
		},
		Key: "test_key",
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, provider, "connect etcd failed")

	var k = koanf.New(".")
	if err := k.Load(provider, nil); err != nil {
		t.Fatal(err)
	}
}
