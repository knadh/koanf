package gcpsecretmanager

import (
	"context"
	"net"
	"strings"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type fakeServer struct {
	secretmanagerpb.UnimplementedSecretManagerServiceServer
	parent  string
	filter  string
	secrets []string
	values  map[string]string
	errs    map[string]codes.Code
}

func (f *fakeServer) ListSecrets(_ context.Context, req *secretmanagerpb.ListSecretsRequest) (*secretmanagerpb.ListSecretsResponse, error) {
	f.parent, f.filter = req.GetParent(), req.GetFilter()

	// Serve one secret per page to exercise pagination.
	i := 0
	if req.GetPageToken() != "" {
		for i < len(f.secrets) && f.secrets[i] != req.GetPageToken() {
			i++
		}
	}
	if i >= len(f.secrets) {
		return &secretmanagerpb.ListSecretsResponse{}, nil
	}
	resp := &secretmanagerpb.ListSecretsResponse{
		Secrets: []*secretmanagerpb.Secret{{Name: req.GetParent() + "/secrets/" + f.secrets[i]}},
	}
	if i+1 < len(f.secrets) {
		resp.NextPageToken = f.secrets[i+1]
	}
	return resp, nil
}

func (f *fakeServer) AccessSecretVersion(_ context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	id := strings.TrimSuffix(req.GetName()[strings.Index(req.GetName(), "/secrets/")+len("/secrets/"):], "/versions/latest")
	if c, ok := f.errs[id]; ok {
		return nil, status.Error(c, "fake error")
	}
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name:    req.GetName(),
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(f.values[id])},
	}, nil
}

func newClient(t *testing.T, srv *fakeServer) *secretmanager.Client {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	gs := grpc.NewServer()
	secretmanagerpb.RegisterSecretManagerServiceServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	client, err := secretmanager.NewClient(context.Background(),
		option.WithEndpoint(lis.Addr().String()),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func newFakeServer() *fakeServer {
	return &fakeServer{
		secrets: []string{"db_host", "db_password", "api-key", "disabled", "empty"},
		values: map[string]string{
			"db_host":     "localhost",
			"db_password": "s3cr3t",
			"api-key":     "abc",
		},
		errs: map[string]codes.Code{
			"disabled": codes.FailedPrecondition,
			"empty":    codes.NotFound,
		},
	}
}

func TestRead(t *testing.T) {
	srv := newFakeServer()
	p := ProviderWithClient(Config{Project: "my-project", Filter: "labels.app=myapp"}, newClient(t, srv))

	mp, err := p.Read()
	require.NoError(t, err)
	assert.Equal(t, "projects/my-project", srv.parent)
	assert.Equal(t, "labels.app=myapp", srv.filter)
	assert.Equal(t, map[string]any{
		"db_host":     "localhost",
		"db_password": "s3cr3t",
		"api-key":     "abc",
	}, mp)
}

func TestReadDelim(t *testing.T) {
	p := ProviderWithClient(Config{Project: "my-project", Delim: "_"}, newClient(t, newFakeServer()))

	k := koanf.New(".")
	require.NoError(t, k.Load(p, nil))
	assert.Equal(t, "localhost", k.String("db.host"))
	assert.Equal(t, "s3cr3t", k.String("db.password"))
	assert.Equal(t, "abc", k.String("api-key"))
}

func TestReadFlatPaths(t *testing.T) {
	p := ProviderWithClient(Config{Project: "my-project", Delim: "_", FlatPaths: true}, newClient(t, newFakeServer()))

	mp, err := p.Read()
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", mp["db_password"])
}

func TestReadCallback(t *testing.T) {
	p := ProviderWithClient(Config{
		Project: "my-project",
		Callback: func(key, value string) (string, any) {
			if !strings.HasPrefix(key, "db_") {
				return "", nil
			}
			return strings.TrimPrefix(key, "db_"), strings.ToUpper(value)
		},
	}, newClient(t, newFakeServer()))

	mp, err := p.Read()
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"host": "LOCALHOST", "password": "S3CR3T"}, mp)
}

func TestReadAccessError(t *testing.T) {
	srv := newFakeServer()
	srv.errs["db_host"] = codes.PermissionDenied
	p := ProviderWithClient(Config{Project: "my-project"}, newClient(t, srv))

	_, err := p.Read()
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestReadNoProject(t *testing.T) {
	_, err := Provider(Config{}).Read()
	require.Error(t, err)
}

func TestReadBytes(t *testing.T) {
	_, err := Provider(Config{Project: "my-project"}).ReadBytes()
	require.Error(t, err)
}
