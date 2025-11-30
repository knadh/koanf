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
func Provider(cfg Config) (*Etcd, error) {
	eCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
	}

	c, err := clientv3.New(eCfg)
	if err != nil {
		return nil, err
	}

	return &Etcd{client: c, cfg: cfg}, nil
}

// ReadBytes is not supported by etcd provider.
func (e *Etcd) ReadBytes() ([]byte, error) {
	return nil, errors.New("etcd provider does not support this method")
}

// Read returns a nested config map.
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

func (e *Etcd) Watch(cb func(event interface{}, err error)) error {
	return e.WatchWithContext(context.Background(), cb)
}

func (e *Etcd) WatchWithContext(ctx context.Context, cb func(event interface{}, err error)) error {
	var w clientv3.WatchChan

	if e.cfg.Prefix {
		w = e.client.Watch(ctx, e.cfg.Key, clientv3.WithPrefix())
	} else {
		w = e.client.Watch(ctx, e.cfg.Key)
	}

	go func() {
		var err error

		for wresp := range w {

			if err = wresp.Err(); err != nil {
				cb(nil, err)
				return
			}

			for _, ev := range wresp.Events {
				cb(ev, nil)
			}
		}

		// no need to check for ctx.Done().
		// reference: If the context "ctx" is canceled or timed out, returned "WatchChan" is closed,
		// and "WatchResponse" from this closed channel has zero events and nil "Err()"
		if err = ctx.Err(); err != nil {
			cb(nil, err)
			return
		}

		cb(nil, errors.New("etcd watcher channel closed"))
	}()

	return nil
}
