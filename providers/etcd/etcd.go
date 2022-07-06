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
	Key string
}

type Etcd struct {
	client	*clientv3.Client
	cfg 	Config
}

func Provider (cfg Config) *Etcd {

	clientCfg := clientv3.Config {
		Endpoints: cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
	}

	newClient, err := clientv3.New(clientCfg)
	if err != nil {
		return nil
	}

	return &Etcd { client: newClient, cfg: cfg }
}

func (etcdHandle *Etcd) ReadBytes() ([]byte, error) {
	return nil, errors.New("etcd provider does not support this method")
}

func (etcdHandle *Etcd) Read() (map[string]interface{}, error) {

	var mp = make(map[string]interface{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)

	var resp *clientv3.GetResponse
	var err error

	if etcdHandle.cfg.Prefix {
		if etcdHandle.cfg.Limit {
			resp, err = etcdHandle.client.Get(ctx, etcdHandle.cfg.Key, clientv3.WithPrefix(), 
				clientv3.WithLimit(etcdHandle.cfg.NLimit))
			if err != nil {
				return nil, err
			}
			cancel()
		} else {
			resp, err = etcdHandle.client.Get(ctx, etcdHandle.cfg.Key, clientv3.WithPrefix())
			if err != nil {
				return nil, err
			}
			cancel()
		}
	} else {
		resp, err = etcdHandle.client.Get(ctx, etcdHandle.cfg.Key)
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
