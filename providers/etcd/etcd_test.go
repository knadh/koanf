package etcd

import (
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Watch must report a closed watch channel to the callback instead of
// silently ending, so that callers know hot-reload has stopped.
func TestWatchReportsClosedChannel(t *testing.T) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:1"},
		DialTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	e := &Etcd{client: c, cfg: Config{Key: "k"}}

	// Closing the client closes every watch channel.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	errs := make(chan error, 1)
	if err := e.Watch(func(_ any, err error) { errs <- err }); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errs:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("callback was never invoked after the watch channel closed")
	}
}
