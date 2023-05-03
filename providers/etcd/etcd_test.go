package etcd

import (
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/mock/mockserver"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	servers, err := mockserver.StartMockServers(1)
	if err != nil {
		t.Fatal(err)
	}
	defer servers.Stop()

	address := servers.Servers[0].Address

	provider, err := Provider(Config{
		ClientConfig: clientv3.Config{
			Endpoints:   []string{address},
			DialTimeout: 30 * time.Second,
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
