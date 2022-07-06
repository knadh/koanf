package etcd

import (
	"time"
	"errors"
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Config struct {
	// etcd endpoints
	Endpoints []string

	// timeout
	DialTimeout time.Duration

	// prefix request option
	Prefix bool

	// limit request option
	Limit bool

	// number of limited pairs
	NLimit int64

	// key, key with prefix, etc.
	Keypath string
}

type Etcd struct {
	client	*clientv3.Client
	cfg 	Config
}

func Provider (cfg Config) *Etcd {

	client_cfg := clientv3.Config {
		Endpoints: cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
	}

	new_client, err := clientv3.New(client_cfg)
	if err != nil {
		return nil
	}

	return &Etcd { client: new_client, cfg: cfg }
}

func (Etcd_handle *Etcd) ReadBytes() ([]byte, error) {
	return nil, errors.New("etcd provider does not support this method")
}

func (Etcd_handle *Etcd) Read() (map[string]interface{}, error) {

	var mp = make(map[string]interface{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)

	var resp *clientv3.GetResponse
	var err error

	if Etcd_handle.cfg.Prefix {
		if Etcd_handle.cfg.Limit {
			resp, err = Etcd_handle.client.Get(ctx, Etcd_handle.cfg.Keypath, clientv3.WithPrefix(), 
				clientv3.WithLimit(Etcd_handle.cfg.NLimit))
			if err != nil {
				return nil, err
			}
			cancel()
		} else {
			resp, err = Etcd_handle.client.Get(ctx, Etcd_handle.cfg.Keypath, clientv3.WithPrefix())
			if err != nil {
				return nil, err
			}
			cancel()
		}
	} else {
		resp, err = Etcd_handle.client.Get(ctx, Etcd_handle.cfg.Keypath)
		if err != nil {
			return nil, err
		}
		cancel()
	}

	for i := 0; i < len(resp.Kvs); i++ {
		mp[string(resp.Kvs[i].Key)] = string(resp.Kvs[i].Value)
	}

	return mp, nil
}























