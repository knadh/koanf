package consul

import (
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

// Config represents the Consul client configuration.
type Config struct {
	// Path of the key to read. If Recurse is true, this is treated
	// as a prefix.
	Key string

	// https://www.consul.io/api-docs/kv#read-key
	// If recurse is true, Consul returns an array of keys.
	// It specifies if the lookup should be recursive and treat
	// Key as a prefix instead of a literal match.
	// This is analogous to: consul kv get -recurse key
	Recurse bool

	// Gets additional metadata about the key in addition to the value such
	// as the ModifyIndex and any flags that may have been set on the key.
	// This is analogous to: consul kv get -detailed key
	Detailed bool

	// Consul client config
	Cfg *api.Config
}

// Consul implements the Consul provider.
type Consul struct {
	client *api.Client
	cfg    Config
}

// Provider returns an instance of the Consul provider.
func Provider(cfg Config) (*Consul, error) {
	c, err := api.NewClient(cfg.Cfg)
	if err != nil {
		return nil, err
	}

	return &Consul{client: c, cfg: cfg}, nil
}

// ReadBytes is not supported by the Consul provider.
func (c *Consul) ReadBytes() ([]byte, error) {
	return nil, errors.New("consul provider does not support this method")
}

// Read reads configuration from the Consul provider.
func (c *Consul) Read() (map[string]any, error) {
	var (
		mp = make(map[string]any)
		kv = c.client.KV()
	)

	if c.cfg.Recurse {
		pairs, _, err := kv.List(c.cfg.Key, nil)
		if err != nil {
			return nil, err
		}

		// Detailed information can be obtained using standard koanf flattened delimited keys:
		// For example:
		// "parent1.CreateIndex"
		// "parent1.Flags"
		// "parent1.LockIndex"
		// "parent1.ModifyIndex"
		// "parent1.Session"
		// "parent1.Value"
		if c.cfg.Detailed {
			for _, pair := range pairs {
				m := make(map[string]any)
				m["CreateIndex"] = fmt.Sprintf("%d", pair.CreateIndex)
				m["Flags"] = fmt.Sprintf("%d", pair.Flags)
				m["LockIndex"] = fmt.Sprintf("%d", pair.LockIndex)
				m["ModifyIndex"] = fmt.Sprintf("%d", pair.ModifyIndex)

				if pair.Session == "" {
					m["Session"] = "-"
				} else {
					m["Session"] = fmt.Sprintf("%s", pair.Session)
				}

				m["Value"] = string(pair.Value)

				mp[pair.Key] = m
			}
		} else {
			for _, pair := range pairs {
				mp[pair.Key] = string(pair.Value)
			}
		}

		return mp, nil
	}

	pair, _, err := kv.Get(c.cfg.Key, nil)
	if err != nil {
		return nil, err
	}

	if c.cfg.Detailed {
		m := make(map[string]any)
		m["CreateIndex"] = fmt.Sprintf("%d", pair.CreateIndex)
		m["Flags"] = fmt.Sprintf("%d", pair.Flags)
		m["LockIndex"] = fmt.Sprintf("%d", pair.LockIndex)
		m["ModifyIndex"] = fmt.Sprintf("%d", pair.ModifyIndex)

		if pair.Session == "" {
			m["Session"] = "-"
		} else {
			m["Session"] = fmt.Sprintf("%s", pair.Session)
		}

		m["Value"] = string(pair.Value)

		mp[pair.Key] = m
	} else {
		mp[pair.Key] = string(pair.Value)
	}

	return mp, nil
}

// Watch watches for changes in the Consul API and triggers a callback.
func (c *Consul) Watch(cb func(event any, err error)) error {
	p := make(map[string]any)

	if c.cfg.Recurse {
		p["type"] = "keyprefix"
		p["prefix"] = c.cfg.Key
	} else {
		p["type"] = "key"
		p["key"] = c.cfg.Key
	}

	plan, err := watch.Parse(p)
	if err != nil {
		return err
	}

	plan.Handler = func(_ uint64, val any) {
		cb(val, nil)
	}

	go func() {
		plan.Run(c.cfg.Cfg.Address)
	}()

	return nil
}
