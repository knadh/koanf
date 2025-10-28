package nats

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/maps"
	"github.com/nats-io/nats.go"
)

type Config struct {
	// nats endpoint (comma separated urls are possible, eg "nats://one, nats://two").
	URL string

	// Optional NATS options: nats.Connect(url, ...options)
	Options []nats.Option

	// Optional JetStream options: nc.JetStream(...options)
	JetStreamOptions []nats.JSOpt

	// Bucket is the Nats KV bucket.
	Bucket string

	// Prefix (optional).
	Prefix string

	// If true, keys will be unflattened by delimiter "." into a nested map
	// So, "a.b.c" results in {"a": {"b": {"c": "value" }}}
	// Prefix will be included
	Unflatten bool
}

// Nats implements the nats config provider.
type Nats struct {
	kv  nats.KeyValue
	cfg Config
}

// Provider returns a provider that takes nats config.
func Provider(cfg Config) (*Nats, error) {
	nc, err := nats.Connect(cfg.URL, cfg.Options...)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream(cfg.JetStreamOptions...)
	if err != nil {
		return nil, err
	}

	kv, err := js.KeyValue(cfg.Bucket)
	if err != nil {
		return nil, err
	}

	return &Nats{kv: kv, cfg: cfg}, nil
}

// ReadBytes is not supported by nats provider.
func (n *Nats) ReadBytes() ([]byte, error) {
	return nil, errors.New("nats provider does not support this method")
}

// Read returns a nested config map.
func (n *Nats) Read() (map[string]any, error) {
	keys, err := n.kv.Keys()
	if err != nil {
		return nil, err
	}

	mp := make(map[string]any)
	for _, key := range keys {
		if !strings.HasPrefix(key, n.cfg.Prefix) {
			continue
		}
		res, err := n.kv.Get(key)
		if err != nil {
			return nil, err
		}
		mp[res.Key()] = string(res.Value())
	}
	if n.cfg.Unflatten {
		return maps.Unflatten(mp, "."), nil
	}

	return mp, nil
}

func (n *Nats) Watch(cb func(event any, err error)) error {
	w, err := n.kv.Watch(fmt.Sprintf("%s.*", n.cfg.Prefix))
	if err != nil {
		return err
	}

	start := time.Now()
	go func(watcher nats.KeyWatcher) {
		for update := range watcher.Updates() {
			// ignore nil events and only callback when the event is new (nats always sends one "old" event)
			if update != nil && update.Created().After(start) {
				cb(update, nil)
			}
		}
	}(w)

	return nil
}
