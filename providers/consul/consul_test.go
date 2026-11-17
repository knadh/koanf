package consul

import (
	"github.com/hashicorp/consul/api"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWatchPreservesConfiguredToken(t *testing.T) {
	tokens := make(chan string, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens <- r.Header.Get("X-Consul-Token")
		w.Header().Set("X-Consul-Index", "1")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"Key":"app","Value":"b2s=","ModifyIndex":1}]`))
	}))
	defer server.Close()
	p, err := Provider(Config{Key: "app", Cfg: &api.Config{Address: server.URL, Token: "configured-secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Read(); err != nil {
		t.Fatal(err)
	}
	if got := <-tokens; got != "configured-secret" {
		t.Fatalf("Read token=%q", got)
	}
	if err := p.Watch(func(any, error) {}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-tokens:
		if got != "configured-secret" {
			t.Fatalf("Watch token=%q; want configured-secret", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no watch request")
	}
}
