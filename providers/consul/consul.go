package consul

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	// key or prefix
	Key string

	// recurse flag
	Recurse bool

	// detailed flag
	Detailed bool
}

type CProvider struct {
	client *api.Client
	cfg Config
}

func Provider (cfg Config) *CProvider {
	
	newClient, err := api.NewClient(api.DefaultConfig())
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
		if cProvider.cfg.Detailed {
			// TODO: detailed info
			return nil, nil
		} else {
			pairs, _, err := kv.List(cProvider.cfg.Key, nil)
			if err != nil {
				return nil, err
			}

			for _, pair := range pairs {
				mp[pair.Key] = string(pair.Value)
			}
		}
	} else {
		if cProvider.cfg.Detailed {
			// TODO: detailed info
			return nil, nil
		} else {
			pair, _, err := kv.Get(cProvider.cfg.Key, nil)
			if err != nil {
				return nil, err
			}

			mp[pair.Key] = string(pair.Value)
		}
	}

	return mp, nil
}
