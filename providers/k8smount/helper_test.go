package k8smount_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	mountTimeFmt = "..2006_01_02_15_04_05.0000000000"
	dataDir      = "..data"
)

// writeVolumeMount creates a file structure that matches how a ConfigMap or Secret will be mounted
// in a Kubernetes Pod.
//
// First, files are created for each data field. These exist within a timestamp-based directory,
// likely when the ConfigMap or Secret was last modified.
//
//	..2025_06_28_09_28_32.3151791122/
//	├── database_hostname
//	├── database_name
//	└── database_port
//
// A symlink is then created for "..data" to the directory containing the data:
//
//	..data -> ..2025_06_28_09_28_32.3151791122
//
// Finally, symlinks are created for the data files, via the "..data" symlink:
//
//	database_hostname -> ..data/database_hostname
//	database_name -> ..data/database_name
//	database_port -> ..data/database_port
func writeVolumeMount(tb testing.TB, mount string, data map[string]string) error {
	tb.Helper()

	return writeVolumeMountAt(tb, time.Now(), mount, data)
}

// writeVolumeMountAtTime creates a file structure that matches how a ConfigMap or Secret will be
// mounted in a Kubernetes Pod, using the given time. See writeVolumeMount for a detailed
// description of the resulting file structure.
//
// This function can be called multiple times with different times, with the most recent call taking
// precedence if any conflicts occur.
func writeVolumeMountAt(tb testing.TB, t time.Time, mount string, data map[string]string) error {
	tb.Helper()

	dir := t.UTC().Format(mountTimeFmt)
	dirPath := filepath.Join(mount, dir)

	for key, value := range data {
		if err := writeFile(tb, filepath.Join(dirPath, key), value); err != nil {
			return err
		}
	}

	tb.Chdir(mount)

	if err := os.Remove(dataDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup existing symlink %s: %w", dataDir, err)
	}

	if err := os.Symlink(dir, dataDir); err != nil {
		return fmt.Errorf("failed to create %s symlink to %q: %w", "..data", dir, err)
	}

	tb.Log("created symlink: ..data ->", dir)

	for key := range data {
		target := filepath.Join(dataDir, key)

		if err := os.Remove(key); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to cleanup existing symlink %s: %w", key, err)
		}

		if err := os.Symlink(target, key); err != nil {
			return fmt.Errorf("failed to create %q symlink to %q: %w", key, target, err)
		}

		tb.Log("created symlink:", key, "->", target)
	}

	return nil
}

func writeFile(tb testing.TB, path, content string) error {
	tb.Helper()

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}

	tb.Log("wrote:", path)

	return nil
}
