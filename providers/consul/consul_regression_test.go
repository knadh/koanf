package consul

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
)

type mockTransport struct {
	status int
	body   string
}

func (m mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: m.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(m.body)), Request: r}, nil
}

func TestReadMissingKey(t *testing.T) {
	for _, detailed := range []bool{false, true} {
		for _, tt := range []struct {
			name    string
			status  int
			body    string
			wantErr bool
			wantLen int
		}{
			{"not_found", 404, "", false, 0},
			{"empty_array", 200, "[]", false, 0},
			{"existing", 200, `[{"Key":"key","Value":"dmFsdWU="}]`, false, 1},
			{"server_error", 500, "server error", true, 0},
		} {
			t.Run(fmt.Sprintf("%s/detailed=%v", tt.name, detailed), func(t *testing.T) {
				p, err := Provider(Config{Key: "key", Detailed: detailed, Cfg: &api.Config{Address: "localhost:8500", HttpClient: &http.Client{Transport: mockTransport{tt.status, tt.body}}}})
				if err != nil {
					t.Fatal(err)
				}
				got, err := p.Read()
				if (err != nil) != tt.wantErr {
					t.Fatalf("error=%v wantErr=%v", err, tt.wantErr)
				}
				if len(got) != tt.wantLen {
					t.Fatalf("got %#v", got)
				}
				if !tt.wantErr && got == nil {
					t.Fatal("expected non-nil map")
				}
				if tt.wantLen == 1 {
					if detailed {
						if got["key"].(map[string]any)["Value"] != "value" {
							t.Fatalf("got %#v", got)
						}
					} else if got["key"] != "value" {
						t.Fatalf("got %#v", got)
					}
				}
			})
		}
	}
}
