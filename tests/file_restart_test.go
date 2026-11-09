package koanf_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/require"
)

func TestWatchFileImmediateRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{}`), 0600))
	f := file.Provider(path)
	t.Cleanup(func() { require.NoError(t, f.Unwatch()) })

	errors := make(chan error, 100)
	for i := 0; i < 100; i++ {
		require.NoError(t, f.Watch(func(_ any, err error) {
			if err != nil {
				select {
				case errors <- err:
				default:
				}
			}
		}))
		require.NoError(t, f.Unwatch())
	}

	changes := make(chan error, 1)
	require.NoError(t, f.Watch(func(_ any, err error) {
		select {
		case changes <- err:
		default:
		}
	}))
	require.NoError(t, os.WriteFile(path, []byte(`{"changed":true}`), 0600))
	select {
	case err := <-changes:
		require.NoError(t, err)
	case err := <-errors:
		t.Fatalf("retired watcher reported an error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("replacement watcher did not receive the file change")
	}
	select {
	case err := <-errors:
		t.Fatalf("retired watcher reported an error: %v", err)
	default:
	}
}
