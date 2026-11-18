package koanf_test

import (
	"reflect"
	"testing"

	"github.com/knadh/koanf/maps"
)

// sliceElems returns slice elements as []any without requiring a concrete []any type.
func sliceElems(t *testing.T, v any) []any {
	t.Helper()
	if v == nil {
		t.Fatal("got nil, want slice")
	}
	if s, ok := v.([]any); ok {
		return s
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		t.Fatalf("want slice, got %T (%#v)", v, v)
	}
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out
}

func assertSliceEqual(t *testing.T, got any, want []any) {
	t.Helper()
	if !reflect.DeepEqual(sliceElems(t, got), want) {
		t.Fatalf("slice=%#v (%T), want %#v", got, got, want)
	}
}

func TestMergeStrategyAppendTypedAndNested(t *testing.T) {
	dest := map[string]any{
		"tags":     []string{"a", "b"},
		"nested":   map[string]any{"keep": 1, "list": []any{1}},
		"scalar":   "old",
		"onlyDest": true,
	}
	src := map[string]any{
		"tags":    []string{"c"},
		"nested":  map[string]any{"add": 2, "list": []any{2}},
		"scalar":  "new",
		"onlySrc": 9,
	}
	if err := maps.MergeAppendSlices(src, dest); err != nil {
		t.Fatal(err)
	}

	assertSliceEqual(t, dest["tags"], []any{"a", "b", "c"})

	nested := dest["nested"].(map[string]any)
	if nested["keep"] != 1 || nested["add"] != 2 {
		t.Fatalf("nested maps not merged: %#v", nested)
	}
	assertSliceEqual(t, nested["list"], []any{1, 2})
	if dest["scalar"] != "new" {
		t.Fatalf("scalar = %v, want new", dest["scalar"])
	}
	if dest["onlyDest"] != true || dest["onlySrc"] != 9 {
		t.Fatalf("key union failed: %#v", dest)
	}
}

func TestMergeStrategyAppendEmptyDest(t *testing.T) {
	dest := map[string]any{"tags": []any{}}
	src := map[string]any{"tags": []any{"x"}}
	if err := maps.MergeAppendSlices(src, dest); err != nil {
		t.Fatal(err)
	}
	assertSliceEqual(t, dest["tags"], []any{"x"})
}

func TestMergeStrategyByIndexMapsAndScalars(t *testing.T) {
	dest := map[string]any{
		"providers": []any{
			map[string]any{"id": "google", "client_id": 1234},
			map[string]any{"id": "okta", "client_id": 1},
		},
		"flags": []string{"a"},
	}
	src := map[string]any{
		"providers": []any{
			map[string]any{"client_secret": "xxxx"},
			map[string]any{"client_id": 2},
			map[string]any{"id": "azure"},
		},
		"flags": []string{"b", "c"},
	}
	if err := maps.MergeByIndex(src, dest); err != nil {
		t.Fatal(err)
	}

	providers := sliceElems(t, dest["providers"])
	if len(providers) != 3 {
		t.Fatalf("len(providers)=%d, want 3", len(providers))
	}
	p0 := providers[0].(map[string]any)
	if p0["id"] != "google" || p0["client_id"] != 1234 || p0["client_secret"] != "xxxx" {
		t.Fatalf("providers[0]=%#v", p0)
	}
	p1 := providers[1].(map[string]any)
	if p1["id"] != "okta" || p1["client_id"] != 2 {
		t.Fatalf("providers[1]=%#v", p1)
	}
	p2 := providers[2].(map[string]any)
	if p2["id"] != "azure" {
		t.Fatalf("providers[2]=%#v", p2)
	}

	assertSliceEqual(t, dest["flags"], []any{"b", "c"})
}

func TestMergeStrategyByKeyProviders(t *testing.T) {
	dest := map[string]any{
		"providers": []any{
			map[string]any{"id": "google", "client_id": 1234},
			map[string]any{"id": "okta", "client_id": 1},
		},
	}
	src := map[string]any{
		"providers": []any{
			map[string]any{"id": "google", "client_secret": "xxxx"},
			map[string]any{"id": "azure", "client_id": 9},
		},
	}
	if err := maps.MergeByKey("id")(src, dest); err != nil {
		t.Fatal(err)
	}
	providers := sliceElems(t, dest["providers"])
	if len(providers) != 3 {
		t.Fatalf("len=%d, want 3", len(providers))
	}
	p0 := providers[0].(map[string]any)
	if p0["id"] != "google" || p0["client_id"] != 1234 || p0["client_secret"] != "xxxx" {
		t.Fatalf("google merge failed: %#v", p0)
	}
	if providers[1].(map[string]any)["id"] != "okta" {
		t.Fatalf("okta missing: %#v", providers[1])
	}
	if providers[2].(map[string]any)["id"] != "azure" {
		t.Fatalf("azure not appended: %#v", providers[2])
	}
}

func TestMergeStrategyByKeyRejectsEmptyField(t *testing.T) {
	err := maps.MergeByKey("")(map[string]any{"a": 1}, map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestMergeStrategyByKeyNestedRoutes(t *testing.T) {
	dest := map[string]any{
		"routes": []any{
			map[string]any{"path": "/v1", "methods": []string{"GET"}},
		},
	}
	src := map[string]any{
		"routes": []any{
			map[string]any{"path": "/v1", "methods": []string{"POST"}},
			map[string]any{"path": "/v2", "methods": []string{"GET"}},
		},
	}
	if err := maps.MergeByKey("path")(src, dest); err != nil {
		t.Fatal(err)
	}
	routes := sliceElems(t, dest["routes"])
	if len(routes) != 2 {
		t.Fatalf("len=%d want 2", len(routes))
	}
	r0 := routes[0].(map[string]any)
	assertSliceEqual(t, r0["methods"], []any{"POST"})
	if routes[1].(map[string]any)["path"] != "/v2" {
		t.Fatalf("missing /v2: %#v", routes)
	}
}

// Nested slices of maps must keep using the same by-key field recursively.
// A shallow deep-merge that overwrites nested map-slices fails this.
func TestMergeStrategyByKeyRecursiveNestedMapSlices(t *testing.T) {
	dest := map[string]any{
		"services": []any{
			map[string]any{
				"id": "api",
				"backends": []any{
					map[string]any{"id": "b1", "host": "old"},
				},
			},
		},
	}
	src := map[string]any{
		"services": []any{
			map[string]any{
				"id": "api",
				"backends": []any{
					map[string]any{"id": "b1", "port": 8080},
					map[string]any{"id": "b2", "host": "new"},
				},
			},
		},
	}
	if err := maps.MergeByKey("id")(src, dest); err != nil {
		t.Fatal(err)
	}

	services := sliceElems(t, dest["services"])
	if len(services) != 1 {
		t.Fatalf("len(services)=%d, want 1", len(services))
	}
	backends := sliceElems(t, services[0].(map[string]any)["backends"])
	if len(backends) != 2 {
		t.Fatalf("len(backends)=%d, want 2 (matched b1 + appended b2)", len(backends))
	}
	b1 := backends[0].(map[string]any)
	if b1["id"] != "b1" || b1["host"] != "old" || b1["port"] != 8080 {
		t.Fatalf("nested by-key merge failed for b1: %#v", b1)
	}
	if backends[1].(map[string]any)["id"] != "b2" {
		t.Fatalf("nested by-key did not append b2: %#v", backends[1])
	}
}

func TestMergeStrategyDefaultMapsOverwrite(t *testing.T) {
	dest := map[string]any{"tags": []any{"a"}}
	src := map[string]any{"tags": []any{"b"}}
	maps.Merge(src, dest)
	assertSliceEqual(t, dest["tags"], []any{"b"})
}

// By-index must recurse into nested slices, merging their elements by index
// as well, not overwrite the whole inner slice.
func TestMergeStrategyByIndexNestedSlices(t *testing.T) {
	dest := map[string]any{
		"grid": []any{
			[]any{1, 2, 3},
			[]string{"a"},
		},
	}
	src := map[string]any{
		"grid": []any{
			[]any{9},
			[]string{"b", "c"},
			[]any{0},
		},
	}
	if err := maps.MergeByIndex(src, dest); err != nil {
		t.Fatal(err)
	}
	grid := sliceElems(t, dest["grid"])
	if len(grid) != 3 {
		t.Fatalf("len(grid)=%d, want 3", len(grid))
	}
	assertSliceEqual(t, grid[0], []any{9, 2, 3})
	assertSliceEqual(t, grid[1], []any{"b", "c"})
	assertSliceEqual(t, grid[2], []any{0})
}

// When several destination elements share the field value, the first one is
// the match; later duplicates stay untouched.
func TestMergeStrategyByKeyFirstMatchWins(t *testing.T) {
	dest := map[string]any{
		"providers": []any{
			map[string]any{"id": "google", "n": 1},
			map[string]any{"id": "google", "n": 2},
		},
	}
	src := map[string]any{
		"providers": []any{
			map[string]any{"id": "google", "client_secret": "xxxx"},
		},
	}
	if err := maps.MergeByKey("id")(src, dest); err != nil {
		t.Fatal(err)
	}
	providers := sliceElems(t, dest["providers"])
	if len(providers) != 2 {
		t.Fatalf("len=%d, want 2", len(providers))
	}
	p0 := providers[0].(map[string]any)
	if p0["n"] != 1 || p0["client_secret"] != "xxxx" {
		t.Fatalf("first duplicate not merged: %#v", p0)
	}
	p1 := providers[1].(map[string]any)
	if p1["n"] != 2 {
		t.Fatalf("second duplicate changed: %#v", p1)
	}
	if _, leaked := p1["client_secret"]; leaked {
		t.Fatalf("second duplicate must stay untouched: %#v", p1)
	}
}

// Merged-in values must be copies: mutating the source afterwards must not
// change what was merged into the destination.
func TestMergeStrategyAppendDoesNotAliasSource(t *testing.T) {
	srcItem := map[string]any{"id": "b1", "host": "orig"}
	src := map[string]any{
		"backends": []any{srcItem},
		"nested":   map[string]any{"list": []any{[]any{"x"}}},
	}
	dest := map[string]any{
		"backends": []any{map[string]any{"id": "b0"}},
	}
	if err := maps.MergeAppendSlices(src, dest); err != nil {
		t.Fatal(err)
	}

	srcItem["host"] = "mutated"
	src["nested"].(map[string]any)["list"].([]any)[0].([]any)[0] = "mutated"

	backends := sliceElems(t, dest["backends"])
	if len(backends) != 2 {
		t.Fatalf("len(backends)=%d, want 2", len(backends))
	}
	merged := backends[1].(map[string]any)
	if merged["host"] != "orig" {
		t.Fatalf("dest aliases source map: %#v", merged)
	}
	inner := sliceElems(t, sliceElems(t, dest["nested"].(map[string]any)["list"])[0])
	if inner[0] != "x" {
		t.Fatalf("dest aliases nested source slice: %#v", inner)
	}
}
