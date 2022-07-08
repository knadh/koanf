package etcd

import (
	"context"
	"errors"
	"time"

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

// Etcd implements the etcd config provider.
type Etcd struct {
	client *clientv3.Client
	cfg    Config
}

// Provider returns a provider that takes etcd config.
func Provider(cfg Config) *Etcd {
	eCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
	}

	c, err := clientv3.New(eCfg)
	if err != nil {
		return nil
	}

	return &Etcd{client: c, cfg: cfg}
}

func (e *Etcd) ReadBytes() ([]byte, error) {
	return nil, errors.New("etcd provider does not support this method")
}

func (e *Etcd) Read() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.DialTimeout)
	defer cancel()

	var resp *clientv3.GetResponse
	if e.cfg.Prefix {
		if e.cfg.Limit {
			r, err := e.client.Get(ctx, e.cfg.Key, clientv3.WithPrefix(), clientv3.WithLimit(e.cfg.NLimit))
			if err != nil {
				return nil, err
			}

			resp = r
		} else {
			r, err := e.client.Get(ctx, e.cfg.Key, clientv3.WithPrefix())
			if err != nil {
				return nil, err
			}

			resp = r
		}
	} else {
		r, err := e.client.Get(ctx, e.cfg.Key)
		if err != nil {
			return nil, err
		}

		resp = r
	}

	mp := make(map[string]interface{}, len(resp.Kvs))
	for _, r := range resp.Kvs {
		mp[string(r.Key)] = string(r.Value)
	}

	return mp, nil
}
