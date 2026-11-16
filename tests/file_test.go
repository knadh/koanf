package koanf_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchFile(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	// Create a tmp config file.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "koanf_mock")
	err := os.WriteFile(tmpFile, []byte(`{"parent": {"name": "name1"}}`), 0600)
	require.NoError(t, err, "error creating temp config file: %v", err)

	// Load the new config and watch it for changes.
	f := file.Provider(tmpFile)
	require.NoError(t, k.Load(f, json.Parser()))

	// Watch for changes.
	changedC := make(chan error, 1)
	require.NoError(t, f.Watch(func(event any, err error) {
		if err == nil {
			// Reload the config.
			err = k.Load(f, json.Parser())
		}

		select {
		case changedC <- err:
		default:
		}
	}))
	defer f.Unwatch()

	// Replace the file atomically so fsnotify cannot report a write while the
	// watched file is truncated but not fully rewritten yet.
	tmpWriteFile := filepath.Join(tmpDir, "koanf_mock.next")
	require.NoError(t, os.WriteFile(tmpWriteFile, []byte(`{"parent": {"name": "name2"}}`), 0600))
	require.NoError(t, os.Rename(tmpWriteFile, tmpFile))

	select {
	case err := <-changedC:
		require.NoError(t, err)
		assert.Equal("name2", k.String("parent.name"), "file watch reload didn't change config")
	case <-time.After(10 * time.Second):
		assert.Fail("timeout waiting for file watch trigger")
	}
}

func TestWatchMissingFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "koanf_mock")
	f := file.Provider(tmpFile)

	// Watching a file that doesn't exist yet should fail.
	require.Error(t, f.Watch(func(event any, err error) {}))

	// The failed Watch() should not leave the provider locked. Retrying once
	// the file exists and unwatching should both return.
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"parent": {"name": "name1"}}`), 0600))

	done := make(chan error, 1)
	go func() {
		if err := f.Watch(func(event any, err error) {}); err != nil {
			done <- err
			return
		}
		done <- f.Unwatch()
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Watch() blocked after a previous Watch() failed")
	}
}

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

func TestWatchFileSymlink(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	tmpDir, _ := os.MkdirTemp("", "koanf_*") // TODO: replace with t.TempDir() as of go v1.15

	// Create a symlink.
	symPath := filepath.Join(tmpDir, "koanf_test_symlink")
	symPath2 := filepath.Join(tmpDir, "koanf_test_symlink2")

	wd, err := os.Getwd()
	assert.NoError(err, "error getting working dir")

	jsonFile := filepath.Join(wd, mockJSON)
	yamlFile := filepath.Join(wd, mockYAML)

	// Create a symlink to the JSON file which will be swapped out later.
	assert.NoError(os.Symlink(jsonFile, symPath), "error creating symlink")
	assert.NoError(os.Symlink(yamlFile, symPath2), "error creating symlink2")

	// Load the symlink (to the JSON) file.
	f := file.Provider(symPath)
	k.Load(f, json.Parser())

	// Watch for changes.
	changedC := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1) // our assurance that cb is called max once
	f.Watch(func(event any, err error) {
		// The File watcher always returns a nil `event`, which can
		// be ignored.
		if err != nil {
			// TODO: make use of Error Wrapping-Scheme and assert.ErrorIs() checks as of go v1.13
			assert.Condition(func() bool {
				return strings.Contains(err.Error(), "no such file or directory")
			}, "received unexpected error. err: %s", err)
			return
		}
		// Reload the config.
		k.Load(f, yaml.Parser())
		changedC <- k.String("type")
		wg.Done()
	})

	// Wait a second and swap the symlink target from the JSON file to the YAML file.
	// Create a temp symlink to the YAML file and rename the old symlink to the new
	// symlink. We do this to avoid removing the symlink and triggering a REMOVE event.
	time.Sleep(1 * time.Second)
	assert.NoError(os.Rename(symPath2, symPath), "error swapping symlink to another file type")
	if waitTimeout(&wg, time.Second*10) {
		assert.Fail("timeout waiting for file watch trigger")
	} else {
		assert.Condition(func() bool {
			return strings.Compare(<-changedC, "yml") == 0
		}, "symlink watch reload didn't change config")
	}
}

func TestWatchFileDirectorySymlink(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	tmpDir := t.TempDir()

	baseDir := filepath.Join(tmpDir, "base_dir")
	baseDir2 := filepath.Join(tmpDir, "base_dir2")

	err := os.Mkdir(baseDir, 0700)
	assert.NoError(err, "error creating base dir")

	err = os.Mkdir(baseDir2, 0700)
	assert.NoError(err, "error creating base dir 2")

	wd, err := os.Getwd()
	assert.NoError(err, "error getting working dir")

	jsonFile := filepath.Join(wd, mockJSON)
	yamlFile := filepath.Join(wd, mockYAML)

	jsonData, err := os.ReadFile(jsonFile)
	assert.NoError(err, "error reading JSON file")

	err = os.WriteFile(filepath.Join(baseDir, "config"), jsonData, 0600)
	assert.NoError(err, "error writing JSON file to base dir")

	yamlData, err := os.ReadFile(yamlFile)
	assert.NoError(err, "error reading YAML file")

	err = os.WriteFile(filepath.Join(baseDir2, "config"), yamlData, 0600)
	assert.NoError(err, "error writing YAML file to base dir 2")

	// Create a symlink.
	symDir := filepath.Join(tmpDir, "koanf_test_symlink")
	symDir2 := filepath.Join(tmpDir, "koanf_test_symlink2")
	symPath := filepath.Join(tmpDir, "config")

	// Create a symlink to the JSON file which will be swapped out later.
	assert.NoError(os.Symlink(baseDir, symDir), "error creating symlink dir")
	assert.NoError(os.Symlink(baseDir2, symDir2), "error creating symlink dir2")
	assert.NoError(os.Symlink(filepath.Join(symDir, "config"), symPath), "error creating symlink")

	// Load the symlink (to the JSON) file.
	f := file.Provider(symPath)
	k.Load(f, json.Parser())

	// Watch for changes.
	changedC := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1) // our assurance that cb is called max once
	f.Watch(func(event any, err error) {
		// The File watcher always returns a nil `event`, which can
		// be ignored.
		if err != nil {
			// TODO: make use of Error Wrapping-Scheme and assert.ErrorIs() checks as of go v1.13
			assert.Condition(func() bool {
				return strings.Contains(err.Error(), "no such file or directory")
			}, "received unexpected error. err: %s", err)
			return
		}
		// Reload the config.
		k.Load(f, yaml.Parser())
		changedC <- k.String("type")
		wg.Done()
	})

	// Wait a second and swap the symlink target from the JSON file to the YAML file.
	// Create a temp symlink to the YAML file and rename the old symlink to the new
	// symlink. We do this to avoid removing the symlink and triggering a REMOVE event.
	time.Sleep(1 * time.Second)
	assert.NoError(os.Rename(symDir2, symDir), "error swapping symlink dir to another symlink dir")

	if waitTimeout(&wg, time.Second*10) {
		assert.Fail("timeout waiting for file watch trigger")
	} else {
		assert.Condition(func() bool {
			return strings.Compare(<-changedC, "yml") == 0
		}, "symlink watch reload didn't change config")
	}
}

func TestUnwatchFile(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	// Create a tmp config file.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "koanf_mock")
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"parent": {"name": "name1"}}`), 0600))

	// Load the new config file.
	f := file.Provider(tmpFile)
	k.Load(f, json.Parser())

	// Watch.
	var reloaded int32
	f.Watch(func(event any, err error) {
		atomic.StoreInt32(&reloaded, 1)
		assert.NoError(err)
	})

	// Change the file and check whether the watch triggered.
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(tmpFile, []byte(`{"parent": {"name": "name2"}}`), 0600)
	time.Sleep(100 * time.Millisecond)
	assert.True(atomic.LoadInt32(&reloaded) == 1, "watched file didn't reload")

	// Unwatch the file and verify that the watch didn't trigger.
	assert.NoError(f.Unwatch())
	atomic.StoreInt32(&reloaded, 0)
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(tmpFile, []byte(`{"parent": {"name": "name3"}}`), 0600)
	time.Sleep(100 * time.Millisecond)
	assert.False(atomic.LoadInt32(&reloaded) == 1, "unwatched file reloaded")

	// Re-watch and check again.
	atomic.StoreInt32(&reloaded, 0)
	f.Watch(func(event any, err error) {
		atomic.StoreInt32(&reloaded, 1)
		assert.NoError(err)
	})
	os.WriteFile(tmpFile, []byte(`{"parent": {"name": "name4"}}`), 0600)
	time.Sleep(100 * time.Millisecond)
	assert.True(atomic.LoadInt32(&reloaded) == 1, "watched file didn't reload")

	f.Unwatch()
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// TestFileWatcherRaceCondition reproduces Issue #305
// File watcher reloading config while reader goroutine accesses values
// This test verifies our thread safety fix prevents empty string reads
func TestFileWatcherRaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	// Create temp config file
	tmpDir, err := os.MkdirTemp("", "koanf_race_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "config.yaml")
	writeConfig := func(value string) {
		content := fmt.Sprintf("rpc: %q\n", value)
		os.WriteFile(tmpFile, []byte(content), 0600)
	}

	// Initial config
	writeConfig("initial")

	k := koanf.New(".")
	provider := file.Provider(tmpFile)
	err = k.Load(provider, yaml.Parser())
	if err != nil {
		t.Fatal(err)
	}

	// Setup file watcher that reloads the SAME koanf instance
	// This tests our internal thread safety, not user pattern issues
	var reloadCount int64
	provider.Watch(func(event any, err error) {
		if err != nil {
			t.Logf("watch error: %v", err)
			return
		}

		// Reload into the same koanf instance - this tests our thread safety
		err = k.Load(provider, yaml.Parser())
		if err != nil {
			t.Logf("reload error: %v", err)
		}
		// Use atomic to avoid race in test counters
		atomic.AddInt64(&reloadCount, 1)
	})

	// Start reader goroutine that continuously reads the config
	done := make(chan struct{})
	var emptyCount, totalReads int64

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				value := k.String("rpc")
				reads := atomic.AddInt64(&totalReads, 1)
				if value == "" {
					atomic.AddInt64(&emptyCount, 1)
					// With proper thread safety, this should never happen
					t.Errorf("Got empty string on read #%d", reads)
				}
				time.Sleep(time.Microsecond) // Small delay to allow interleaving
			}
		}
	}()

	// Trigger multiple file changes to increase chance of race
	for i := 0; i < 5; i++ {
		time.Sleep(20 * time.Millisecond)
		writeConfig(fmt.Sprintf("value-%d", i))
	}

	// Wait for file watching to settle
	time.Sleep(100 * time.Millisecond)
	close(done)

	finalReads := atomic.LoadInt64(&totalReads)
	finalEmpties := atomic.LoadInt64(&emptyCount)
	finalReloads := atomic.LoadInt64(&reloadCount)

	t.Logf("Total reads: %d, Empty reads: %d, Reloads: %d", finalReads, finalEmpties, finalReloads)
	if finalEmpties > 0 {
		t.Errorf("Thread safety issue: got %d empty reads out of %d total reads", finalEmpties, finalReads)
	}
}

// TestFileProviderConcurrency specifically tests the file provider's synchronization
// under heavy concurrent load to catch any races in Watch/Unwatch
func TestFileProviderConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping file provider concurrency test in short mode")
	}

	// Create temp config file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")
	initialData := []byte(`{"test": "value"}`)
	err := os.WriteFile(tmpFile, initialData, 0600)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Stress test: multiple goroutines rapidly watch/unwatch the same file
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-done:
					return
				default:
					// Create new file provider for each iteration
					f := file.Provider(tmpFile)

					// Try to watch
					var watchErr error
					watchErr = f.Watch(func(event any, err error) {
						// Simple callback that doesn't do much
						_ = event
						_ = err
					})

					// Sometimes watch will fail if another goroutine is already watching
					if watchErr == nil {
						// If watch succeeded, unwatch after a short delay
						time.Sleep(time.Millisecond)
						unwatchErr := f.Unwatch()
						if unwatchErr != nil {
							// Log but don't fail - this can happen during cleanup
							t.Logf("Unwatch error (normal during stress test): %v", unwatchErr)
						}
					} else if watchErr.Error() != "file is already being watched" {
						// Unexpected error
						t.Errorf("Unexpected watch error: %v", watchErr)
					}

					// Small delay to prevent tight loop
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	// Also test concurrent file modifications while watching
	wg.Add(1)
	go func() {
		defer wg.Done()
		counter := 0

		for {
			select {
			case <-done:
				return
			default:
				// Write new data to trigger file events
				newData := fmt.Sprintf(`{"test": "value", "counter": %d}`, counter)
				os.WriteFile(tmpFile, []byte(newData), 0600)
				counter++
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Let the stress test run for 200ms
	time.Sleep(200 * time.Millisecond)
	close(done)

	// Wait with timeout to detect deadlocks
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		t.Log("File provider concurrency test completed successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("FILE PROVIDER DEADLOCK: Goroutines did not complete within timeout")
	}
}
