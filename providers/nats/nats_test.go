//go:build go1.19
// +build go1.19

package nats

import (
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestNats(t *testing.T) {
	k := koanf.NewWithConf(koanf.Conf{})

	nc, err := nats.Connect(testNatsURL)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "test",
	})
	_, err = kv.Put("some.test.color", []byte("blue"))
	if err != nil {
		t.Fatal(err)
	}

	provider, err := Provider(Config{
		URL:    testNatsURL,
		Bucket: "test",
		Prefix: "some.test",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = k.Load(provider, nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, k.Keys(), []string{"some.test.color"})
	assert.Equal(t, k.Get("some.test.color"), "blue")

	err = provider.Watch(func(event any, err error) {
		if err != nil {
			t.Fatal(err)
		}

		err = k.Load(provider, nil)
		if err != nil {
			t.Fatal(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err := kv.Put("some.test.color", []byte("yellow"))
		if err != nil {
			t.Error(err)
			return
		}
	}()

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, k.Get("some.test.color"), "yellow")
}

func TestWatchWithoutPrefix(t *testing.T) {
	nc, err := nats.Connect(testNatsURL)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "watch_without_prefix"})
	if err != nil {
		t.Fatal(err)
	}
	defer js.DeleteKeyValue("watch_without_prefix")
	if _, err := kv.Put("app.name", []byte("before")); err != nil {
		t.Fatal(err)
	}
	provider, err := Provider(Config{URL: testNatsURL, Bucket: "watch_without_prefix"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := provider.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "before", data["app.name"])
	updates := make(chan nats.KeyValueEntry, 1)
	if err := provider.Watch(func(event any, err error) {
		if err == nil {
			updates <- event.(nats.KeyValueEntry)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := kv.Put("app.name", []byte("after")); err != nil {
		t.Fatal(err)
	}
	select {
	case update := <-updates:
		assert.Equal(t, "app.name", update.Key())
		assert.Equal(t, "after", string(update.Value()))
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for update without a prefix")
	}
}
