package consul

import (
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	// key or prefix
	Key string

	// recurse flag
	Recurse bool

	// detailed flag
	Detailed bool

	// Consul client config
	CConfig *api.Config
}

type CProvider struct {
	client *api.Client
	cfg Config
}

func Provider (cfg Config) *CProvider {
	
	newClient, err := api.NewClient(cfg.CConfig)
	if err != nil {
		return nil
	}

	return &CProvider { client: newClient, cfg: cfg }
}

func (cProvider *CProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("consul provider does not support this method")
}

func (cProvider *CProvider) Read() (map[string]interface{}, error) {
	var mp = make(map[string]interface{})

	kv := cProvider.client.KV()

	if cProvider.cfg.Recurse {
		pairs, _, err := kv.List(cProvider.cfg.Key, nil)
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
		if cProvider.cfg.Detailed {
			for _, pair := range pairs {
				key_meta := make(map[string]interface{})
				key_meta["CreateIndex"] = fmt.Sprintf("%d", pair.CreateIndex)
				key_meta["Flags"] = fmt.Sprintf("%d", pair.Flags)
				key_meta["LockIndex"] = fmt.Sprintf("%d", pair.LockIndex)
				key_meta["ModifyIndex"] = fmt.Sprintf("%d", pair.ModifyIndex)
				if pair.Session == "" {
					key_meta["Session"] = "-"
				} else {
					key_meta["Session"] = fmt.Sprintf("%s", pair.Session)
				}

				key_meta["Value"] = string(pair.Value)

				mp[pair.Key] = key_meta
			}
		} else {
			for _, pair := range pairs {
				mp[pair.Key] = string(pair.Value)
			}
		}
	} else {
		pair, _, err := kv.Get(cProvider.cfg.Key, nil)
		if err != nil {
			return nil, err
		}

		if cProvider.cfg.Detailed {
			key_meta := make(map[string]interface{})
			key_meta["CreateIndex"] = fmt.Sprintf("%d", pair.CreateIndex)
			key_meta["Flags"] = fmt.Sprintf("%d", pair.Flags)
			key_meta["LockIndex"] = fmt.Sprintf("%d", pair.LockIndex)
			key_meta["ModifyIndex"] = fmt.Sprintf("%d", pair.ModifyIndex)
			if pair.Session == "" {
				key_meta["Session"] = "-"
			} else {
				key_meta["Session"] = fmt.Sprintf("%s", pair.Session)
			}

			key_meta["Value"] = string(pair.Value)

			mp[pair.Key] = key_meta
		} else {
			mp[pair.Key] = string(pair.Value)
		}
	}

	return mp, nil
}
