package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(t *testing.T, body string, cfg Config) *Vault {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	cfg.Address = srv.URL
	cfg.Token = "test-token"
	cfg.Path = "secret/data/app"

	p, err := Provider(cfg)
	if err != nil {
		t.Fatalf("Provider: %v", err)
	}

	return p
}

func TestReadExcludeMeta(t *testing.T) {
	p := newTestProvider(t,
		`{"data":{"data":{"key":"value"},"metadata":{"version":1}}}`,
		Config{ExcludeMeta: true})

	got, err := p.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got["key"] != "value" {
		t.Fatalf("unexpected data: %v", got)
	}
}

// A KV v1 secret has no nested "data" map; ExcludeMeta must return an error
// rather than panic on the type assertion.
func TestReadExcludeMetaWithoutNestedData(t *testing.T) {
	p := newTestProvider(t,
		`{"data":{"key":"value"}}`,
		Config{ExcludeMeta: true})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Read panicked: %v", r)
		}
	}()

	if _, err := p.Read(); err == nil {
		t.Fatal("expected an error, got nil")
	}
}
