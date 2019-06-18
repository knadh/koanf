package koanf

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/hcl"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

const (
	delim = "."

	mockDir  = "mock"
	mockJSON = mockDir + "/mock.json"
	mockYAML = mockDir + "/mock.yml"
	mockTOML = mockDir + "/mock.toml"
	mockHCL  = mockDir + "/mock.hcl"
	mockProp = mockDir + "/mock.prop"
)

// Ordered list of fields in the test confs.
var testKeys = []string{
	"bools",
	"empty",
	"intbools",
	"orphan",
	"parent1.child1.empty",
	"parent1.child1.grandchild1.ids",
	"parent1.child1.grandchild1.on",
	"parent1.child1.name",
	"parent1.child1.type",
	"parent1.id",
	"parent1.name",
	"parent2.child2.empty",
	"parent2.child2.grandchild2.ids",
	"parent2.child2.grandchild2.on",
	"parent2.child2.name",
	"parent2.id",
	"parent2.name",
	"strbool",
	"strbools",
	"time",
	"type",
}

var testKeyMap = map[string][]string{
	"bools":                          []string{"bools"},
	"empty":                          []string{"empty"},
	"intbools":                       []string{"intbools"},
	"orphan":                         []string{"orphan"},
	"parent1":                        []string{"parent1"},
	"parent1.child1":                 []string{"parent1", "child1"},
	"parent1.child1.empty":           []string{"parent1", "child1", "empty"},
	"parent1.child1.grandchild1":     []string{"parent1", "child1", "grandchild1"},
	"parent1.child1.grandchild1.ids": []string{"parent1", "child1", "grandchild1", "ids"},
	"parent1.child1.grandchild1.on":  []string{"parent1", "child1", "grandchild1", "on"},
	"parent1.child1.name":            []string{"parent1", "child1", "name"},
	"parent1.child1.type":            []string{"parent1", "child1", "type"},
	"parent1.id":                     []string{"parent1", "id"},
	"parent1.name":                   []string{"parent1", "name"},
	"parent2":                        []string{"parent2"},
	"parent2.child2":                 []string{"parent2", "child2"},
	"parent2.child2.empty":           []string{"parent2", "child2", "empty"},
	"parent2.child2.grandchild2":     []string{"parent2", "child2", "grandchild2"},
	"parent2.child2.grandchild2.ids": []string{"parent2", "child2", "grandchild2", "ids"},
	"parent2.child2.grandchild2.on":  []string{"parent2", "child2", "grandchild2", "on"},
	"parent2.child2.name":            []string{"parent2", "child2", "name"},
	"parent2.id":                     []string{"parent2", "id"},
	"parent2.name":                   []string{"parent2", "name"},
	"strbool":                        []string{"strbool"},
	"strbools":                       []string{"strbools"},
	"time":                           []string{"time"},
	"type":                           []string{"type"},
}

// `parent1.child1.type` and `type` are excluded as their
// values vary between files.
var testAll = `bools -> [true false true]
empty -> map[]
intbools -> [1 0 1]
orphan -> [red blue orange]
parent1.child1.empty -> map[]
parent1.child1.grandchild1.ids -> [1 2 3]
parent1.child1.grandchild1.on -> true
parent1.child1.name -> child1
parent1.id -> 1234
parent1.name -> parent1
parent2.child2.empty -> map[]
parent2.child2.grandchild2.ids -> [4 5 6]
parent2.child2.grandchild2.on -> true
parent2.child2.name -> child2
parent2.id -> 5678
parent2.name -> parent2
strbool -> 1
strbools -> [1 t f]
time -> 2019-01-01`

var testParent2 = `child2.empty -> map[]
child2.grandchild2.ids -> [4 5 6]
child2.grandchild2.on -> true
child2.name -> child2
id -> 5678
name -> parent2`

type parentStruct struct {
	Name   string      `koanf:"name"`
	ID     int         `koanf:"id"`
	Child1 childStruct `koanf:"child1"`
}
type childStruct struct {
	Name        string            `koanf:"name"`
	Type        string            `koanf:"type"`
	Empty       map[string]string `koanf:"empty"`
	Grandchild1 grandchildStruct  `koanf:"grandchild1"`
}
type grandchildStruct struct {
	Ids []int `koanf:"ids"`
	On  bool  `koanf:"on"`
}
type testStruct struct {
	Type    string            `koanf:"type"`
	Empty   map[string]string `koanf:"empty"`
	Parent1 parentStruct      `koanf:"parent1"`
}

type Case struct {
	koanf    *Koanf
	file     string
	parser   Parser
	typeName string
}

// Case instances to be used in multiple tests. These will not be mutated.
var cases = []Case{
	{koanf: New(delim), file: mockJSON, parser: json.Parser(), typeName: "json"},
	{koanf: New(delim), file: mockYAML, parser: yaml.Parser(), typeName: "yml"},
	{koanf: New(delim), file: mockTOML, parser: toml.Parser(), typeName: "toml"},
	{koanf: New(delim), file: mockHCL, parser: hcl.Parser(true), typeName: "hcl"},
}

func init() {
	// Preload 4 Koanf instances with their providers and config.
	if err := cases[0].koanf.Load(file.Provider(cases[0].file), json.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
	if err := cases[1].koanf.Load(file.Provider(cases[1].file), yaml.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
	if err := cases[2].koanf.Load(file.Provider(cases[2].file), toml.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
	if err := cases[3].koanf.Load(file.Provider(cases[3].file), hcl.Parser(true)); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
}

func TestLoadFile(t *testing.T) {
	// Load a non-existent file.
	_, err := file.Provider("does-not-exist").ReadBytes()
	assert.NotNil(t, err, "no error for non-existent file")

	// Load a valid file.
	_, err = file.Provider(mockJSON).ReadBytes()
	assert.Nil(t, err, "error loading file")
}

func TestLoadFileAllKeys(t *testing.T) {
	re, _ := regexp.Compile("(.+?)?type \\-> (.*)\n")
	for _, c := range cases {
		// Check against testKeys.
		assert.Equal(t, testKeys, c.koanf.Keys(), fmt.Sprintf("loaded keys mismatch: %v", c.typeName))

		// Check against keyMap.
		assert.EqualValues(t, testKeyMap, c.koanf.KeyMap(), "keymap doesn't match")

		// Replace the "type" fields that varies across different files
		// to do a complete key -> value map match with testAll.
		s := strings.TrimSpace(re.ReplaceAllString(c.koanf.Sprint(), ""))
		assert.Equal(t, testAll, s, fmt.Sprintf("key -> value list mismatch: %v", c.typeName))
	}
}

func TestLoadMerge(t *testing.T) {
	// Load several types into a fresh Koanf instance.
	k := New(delim)
	for _, c := range cases {
		assert.Nil(t, k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Check against testKeys.
		assert.Equal(t, testKeys, k.Keys(), fmt.Sprintf("loaded keys don't match in: %v", c.file))

		// The 'type' fields in different file types have different values.
		// As each subsequent file is loaded, the previous value should be overridden.
		assert.Equal(t, c.typeName, k.String("type"), "types don't match")
		assert.Equal(t, c.typeName, k.String("parent1.child1.type"), "types don't match")
	}

	// Load env provider and override value.
	os.Setenv("PREFIX_PARENT1.CHILD1.TYPE", "env")
	err := k.Load(env.Provider("PREFIX_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "prefix_", "")
	}), nil)

	assert.Nil(t, err, "error loading env")
	assert.Equal(t, "env", k.String("parent1.child1.type"), "types don't match")

	// Override with the posflag provider.
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("parent1.child1.type", "flag", "")
	f.Set("parent1.child1.type", "posflag")
	assert.Nil(t, k.Load(posflag.Provider(f, "."), nil), "error loading posflag")
	assert.Equal(t, "posflag", k.String("parent1.child1.type"), "types don't match")

	// Override with the flag provider.
	bf := flag.NewFlagSet("test", flag.ContinueOnError)
	bf.String("parent1.child1.type", "flag", "")
	bf.Set("parent1.child1.type", "basicflag")
	assert.Nil(t, k.Load(basicflag.Provider(bf, "."), nil), "error loading basicflag")
	assert.Equal(t, "basicflag", k.String("parent1.child1.type"), "types don't match")

	// Override with the confmap provider.
	k.Load(confmap.Provider(map[string]interface{}{
		"parent1.child1.type": "confmap",
		"type":                "confmap",
	}, "."), nil)
	assert.Equal(t, "confmap", k.String("parent1.child1.type"), "types don't match")
	assert.Equal(t, "confmap", k.String("type"), "types don't match")

	// Override with the rawbytes provider.
	assert.Nil(t,
		k.Load(rawbytes.Provider([]byte(`{"type": "rawbytes", "parent1": {"child1": {"type": "rawbytes"}}}`)), json.Parser()),
		"error loading raw bytes")
	assert.Equal(t, "rawbytes", k.String("parent1.child1.type"), "types don't match")
	assert.Equal(t, "rawbytes", k.String("type"), "types don't match")
}

func TestConfMapValues(t *testing.T) {
	k := New(delim)
	assert.Nil(t, k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	var (
		c1  = k.All()
		ra1 = k.Raw()
		r1  = k.Cut("parent2").Raw()
	)

	k = New(delim)
	assert.Nil(t, k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	var (
		c2  = k.All()
		ra2 = k.Raw()
		r2  = k.Cut("parent2").Raw()
	)

	assert.EqualValues(t, c1, c2, "conf map mismatch")
	assert.EqualValues(t, ra1, ra2, "conf map mismatch")
	assert.EqualValues(t, r1, r2, "conf map mismatch")
}

func TestCutCopy(t *testing.T) {
	// Instance 1.
	k1 := New(delim)
	assert.Nil(t, k1.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	var (
		cp1   = k1.Copy()
		cut1  = k1.Cut("")
		cutp1 = k1.Cut("parent2")
	)

	// Instance 2.
	k2 := New(delim)
	assert.Nil(t, k2.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	var (
		cp2   = k2.Copy()
		cut2  = k2.Cut("")
		cutp2 = k2.Cut("parent2")
	)

	assert.EqualValues(t, cp1.All(), cp2.All(), "conf map mismatch")
	assert.EqualValues(t, cut1.All(), cut2.All(), "conf map mismatch")
	assert.EqualValues(t, cutp1.All(), cutp2.All(), "conf map mismatch")
	assert.Equal(t, testParent2, strings.TrimSpace(cutp1.Sprint()), "conf map mismatch")
	assert.Equal(t, strings.TrimSpace(cutp1.Sprint()), strings.TrimSpace(cutp2.Sprint()), "conf map mismatch")

	// Cut a single field with no children. Should return empty conf maps.
	assert.Equal(t, k1.Cut("type").Keys(), k2.Cut("type").Keys(), "single field cut mismatch")
	assert.Equal(t, k1.Cut("xxxx").Keys(), k2.Cut("xxxx").Keys(), "single field cut mismatch")
	assert.Equal(t, len(k1.Cut("type").Raw()), 0, "non-map cut returned items")
}

func TestMerge(t *testing.T) {
	k := New(delim)
	assert.Nil(t, k.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")

	// Make two different cuts that'll have different confmaps.
	var (
		cut1 = k.Cut("parent1")
		cut2 = k.Cut("parent2")
	)
	assert.NotEqual(t, cut1.All(), cut2.All(), "different cuts incorrectly match")
	assert.NotEqual(t, cut1.Sprint(), cut2.Sprint(), "different cuts incorrectly match")

	// Create an empty Koanf instance.
	k2 := New(delim)

	// Merge cut1 into it and check if they match.
	k2.Merge(cut1)
	assert.Equal(t, cut1.All(), k2.All(), "conf map mismatch")
}

func TestUnmarshal(t *testing.T) {
	// Expected unmarshalled structure.
	real := testStruct{
		Type:  "json",
		Empty: make(map[string]string),
		Parent1: parentStruct{
			Name: "parent1",
			ID:   1234,
			Child1: childStruct{
				Name:  "child1",
				Type:  "json",
				Empty: make(map[string]string),
				Grandchild1: grandchildStruct{
					Ids: []int{1, 2, 3},
					On:  true,
				},
			},
		},
	}

	// Unmarshal and check all parsers.
	for _, c := range cases {
		var (
			k  = New(delim)
			ts testStruct
		)
		assert.Nil(t, k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		assert.Nil(t, k.Unmarshal("", &ts), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(t, real, ts, "unmarshalled structs don't match")

		// Unmarshal with config.
		ts = testStruct{}
		assert.Nil(t, k.UnmarshalWithConf("", &ts, UnmarshalConf{Tag:"koanf"}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(t, real, ts, "unmarshalled structs don't match")
	}
}

func TestGetExists(t *testing.T) {
	type exCase struct {
		path   string
		exists bool
	}
	exCases := []exCase{
		{"xxxxx", false},
		{"parent1.child2", false},
		{"child1", false},
		{"child2", false},
		{"type", true},
		{"parent1", true},
		{"parent2", true},
		{"parent1.name", true},
		{"parent2.name", true},
		{"parent1.child1", true},
		{"parent2.child2", true},
		{"parent1.child1.grandchild1", true},
		{"parent1.child1.grandchild1.on", true},
		{"parent2.child2.grandchild2.on", true},
	}
	for _, c := range exCases {
		assert.Equal(t, c.exists, cases[0].koanf.Exists(c.path),
			fmt.Sprintf("path resolution failed: %s", c.path))
		assert.Equal(t, c.exists, cases[0].koanf.Get(c.path) != nil,
			fmt.Sprintf("path resolution failed: %s", c.path))
	}
}

func TestGetTypes(t *testing.T) {
	for _, c := range cases {
		assert.Equal(t, nil, c.koanf.Get("xxx"), "get value mismatch")
		assert.Equal(t, make(map[string]interface{}), c.koanf.Get("empty"), "get value mismatch")
		assert.Equal(t, int64(0), c.koanf.Int64("xxxx"), "get value mismatch")
		assert.Equal(t, int64(1234), c.koanf.Int64("parent1.id"), "get value mismatch")
		assert.Equal(t, int(0), c.koanf.Int("xxxx"), "get value mismatch")
		assert.Equal(t, int(1234), c.koanf.Int("parent1.id"), "get value mismatch")
		assert.Equal(t, []int64{}, c.koanf.Int64s("xxxx"), "get value mismatch")
		assert.Equal(t, []int64{1, 2, 3}, c.koanf.Int64s("parent1.child1.grandchild1.ids"), "get value mismatch")
		assert.Equal(t, []int{1, 2, 3}, c.koanf.Ints("parent1.child1.grandchild1.ids"), "get value mismatch")
		assert.Equal(t, []int{}, c.koanf.Ints("xxxx"), "get value mismatch")
		assert.Equal(t, float64(0), c.koanf.Float64("xxx"), "get value mismatch")
		assert.Equal(t, float64(1234), c.koanf.Float64("parent1.id"), "get value mismatch")
		assert.Equal(t, []float64{}, c.koanf.Float64s("xxxx"), "get value mismatch")
		assert.Equal(t, []float64{1, 2, 3}, c.koanf.Float64s("parent1.child1.grandchild1.ids"), "get value mismatch")
		assert.Equal(t, []byte{}, c.koanf.Bytes("xxxx"), "get value mismatch")
		assert.Equal(t, []byte("parent1"), c.koanf.Bytes("parent1.name"), "get value mismatch")
		assert.Equal(t, "", c.koanf.String("xxxx"), "get value mismatch")
		assert.Equal(t, "parent1", c.koanf.String("parent1.name"), "get value mismatch")
		assert.Equal(t, []string{}, c.koanf.Strings("xxxx"), "get value mismatch")
		assert.Equal(t, []string{"red", "blue", "orange"}, c.koanf.Strings("orphan"), "get value mismatch")
		assert.Equal(t, false, c.koanf.Bool("xxxx"), "get value mismatch")
		assert.Equal(t, false, c.koanf.Bool("type"), "get value mismatch")
		assert.Equal(t, true, c.koanf.Bool("parent1.child1.grandchild1.on"), "get value mismatch")
		assert.Equal(t, true, c.koanf.Bool("strbool"), "get value mismatch")
		assert.Equal(t, []bool{}, c.koanf.Bools("xxxx"), "get value mismatch")
		assert.Equal(t, []bool{true, false, true}, c.koanf.Bools("bools"), "get value mismatch")
		assert.Equal(t, []bool{true, false, true}, c.koanf.Bools("intbools"), "get value mismatch")
		assert.Equal(t, []bool{true, true, false}, c.koanf.Bools("strbools"), "get value mismatch")
		assert.Equal(t, time.Duration(1234), c.koanf.Duration("parent1.id"), "get value mismatch")
		assert.Equal(t, time.Duration(0), c.koanf.Duration("xxxx"), "get value mismatch")
		assert.Equal(t, time.Time{}, c.koanf.Time("xxxx", "2006-01-02"), "get value mismatch")
		assert.Equal(t, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), c.koanf.Time("time", "2006-01-02"), "get value mismatch")
		// Attempt to parse int=1234 as a Unix timestamp.
		assert.Equal(t, time.Date(1970, 1, 1, 0, 20, 34, 0, time.UTC), c.koanf.Time("parent1.id", "").UTC(), "get value mismatch")
	}
}
