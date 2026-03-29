package k8smount_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/knadh/koanf/providers/k8smount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDelim = "."

func Test_New(t *testing.T) {
	// arrange
	dir := t.TempDir()

	// act
	provider := k8smount.Provider(dir, testDelim, k8smount.Opt{})

	// assert
	assert.NotNil(t, provider)
}

func Test_K8SMount_ReadBytes(t *testing.T) {
	// arrange
	dir := t.TempDir()
	provider := k8smount.Provider(dir, testDelim, k8smount.Opt{})

	// act
	content, err := provider.ReadBytes()

	// assert
	assert.Empty(t, content)
	assert.EqualError(t, err, "k8smount provider does not support this method")
}

func Test_K8SMount_Read_Empty(t *testing.T) {
	// arrange
	dir := t.TempDir()
	provider := k8smount.Provider(dir, testDelim, k8smount.Opt{})

	// act
	values, err := provider.Read()

	// assert
	assert.Empty(t, values)
	assert.NoError(t, err)
}

func Test_K8SMount_Read_WithFiles(t *testing.T) {
	// arrange
	dir := t.TempDir()

	require.NoError(t, writeFile(t, filepath.Join(dir, "a"), "a"))
	require.NoError(t, writeFile(t, filepath.Join(dir, "b.c"), "c"))
	require.NoError(t, writeFile(t, filepath.Join(dir, "d.e.f"), "f"))
	require.NoError(t, writeFile(t, filepath.Join(dir, "g", "h"), "h"))

	provider := k8smount.Provider(dir, testDelim, k8smount.Opt{})

	// act
	got, err := provider.Read()

	// assert
	want := map[string]any{
		"a": "a",
		"b": map[string]any{
			"c": "c",
		},
		"d": map[string]any{
			"e": map[string]any{
				"f": "f",
			},
		},
		"g": map[string]any{
			"h": "h",
		},
	}

	assert.Equal(t, want, got)
	assert.NoError(t, err)
}

func Test_K8SMount_Read_WithVolumeMount(t *testing.T) {
	tests := map[string]struct {
		have          map[string]string
		transformFunc func(k, v string) (string, any)
		want          map[string]any
	}{
		"no transform func": {
			have: map[string]string{
				"a_foo": "foo-value",
				"b_bar": "bar-value",
				"b_baz": "baz-value",
			},
			want: map[string]any{
				"a_foo": "foo-value",
				"b_bar": "bar-value",
				"b_baz": "baz-value",
			},
		},
		"with transform func replace+lowercase": {
			have: map[string]string{
				"a_foo": "foo-value",
				"b_bar": "bar-value",
				"b_baz": "baz-value",
			},
			transformFunc: func(k, v string) (string, any) {
				return strings.ToLower(strings.ReplaceAll(k, "_", testDelim)), v
			},
			want: map[string]any{
				"a": map[string]any{
					"foo": "foo-value",
				},
				"b": map[string]any{
					"bar": "bar-value",
					"baz": "baz-value",
				},
			},
		},
		"with transform func empty string": {
			have: map[string]string{
				"a_foo": "foo-value",
				"b_bar": "bar-value",
				"b_baz": "baz-value",
			},
			transformFunc: func(_, v string) (string, any) {
				return "", v
			},
			want: map[string]any{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			dir := t.TempDir()
			require.NoError(t, writeVolumeMount(t, dir, tt.have))

			provider := k8smount.Provider(dir, "." /*delim*/, k8smount.Opt{
				TransformFunc: tt.transformFunc,
			})

			// act
			got, err := provider.Read()

			// assert
			assert.Equal(t, tt.want, got)
			assert.NoError(t, err)
		})
	}
}

func Test_K8SMount_Read_MissingLink(t *testing.T) {
	// arrange
	now := time.Now()

	dir := t.TempDir()
	require.NoError(t, writeVolumeMountAt(t, now, dir, map[string]string{
		"foo": "foo-value",
	}))

	name := now.UTC().Format(mountTimeFmt)
	file := filepath.Join(dir, name, "foo")
	require.NoError(t, os.Remove(file))

	provider := k8smount.Provider(dir, "." /*delim*/, k8smount.Opt{})

	// act
	got, err := provider.Read()

	// assert
	assert.Empty(t, got)
	assert.NoError(t, err)
}

func Test_K8SMount_Read_MissingDir(t *testing.T) {
	// arrange
	provider := k8smount.Provider("/does/not/exist" /*dir*/, "." /*delim*/, k8smount.Opt{})

	// act
	values, err := provider.Read()

	// assert
	assert.Empty(t, values)
	assert.EqualError(
		t,
		err,
		`failed to open mount: open /does/not/exist: no such file or directory`,
	)
}

func Test_K8SMount_Watch_Success(t *testing.T) {
	// arrange
	dir := t.TempDir()

	require.NoError(t, writeFile(t, filepath.Join(dir, "a"), "a"))

	provider := k8smount.Provider(dir, "." /*delim*/, k8smount.Opt{})

	_, err := provider.Read()
	require.NoError(t, err)

	var watched atomic.Bool

	// act
	require.NoError(t, provider.Watch(func(_ any, err error) {
		assert.NoError(t, err)
		watched.Store(true)
	}))

	for !watched.Load() {
		require.NoError(t, writeFile(t, filepath.Join(dir, "a"), "b"))
	}

	// assert
	require.NoError(t, provider.Unwatch())

	got, err := provider.Read()

	want := map[string]any{
		"a": "b",
	}

	assert.Equal(t, want, got)
	assert.NoError(t, err)
}

func Test_K8SMount_Watch_AlreadyWatching(t *testing.T) {
	// arrange
	dir := t.TempDir()
	provider := k8smount.Provider(dir, "." /*delim*/, k8smount.Opt{})

	require.NoError(t, provider.Watch(func(_ any, err error) {
		assert.NoError(t, err)
	}))
	defer func() {
		assert.NoError(t, provider.Unwatch())
	}()

	// act
	err := provider.Watch(func(_ any, err error) {
		assert.NoError(t, err)
	})

	// assert
	assert.ErrorIs(t, err, k8smount.ErrAlreadyWatched)
}

func Test_K8SMount_Unwatch(t *testing.T) {
	// arrange
	dir := t.TempDir()
	provider := k8smount.Provider(dir, "." /*delim*/, k8smount.Opt{})

	// act
	err := provider.Unwatch()

	// assert
	assert.NoError(t, err)
}
