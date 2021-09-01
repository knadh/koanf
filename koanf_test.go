package koanf_test

import (
	encjson "encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/dotenv"
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
	"github.com/stretchr/testify/require"
)

const (
	delim = "."

	mockDir    = "mock"
	mockJSON   = mockDir + "/mock.json"
	mockYAML   = mockDir + "/mock.yml"
	mockTOML   = mockDir + "/mock.toml"
	mockHCL    = mockDir + "/mock.hcl"
	mockProp   = mockDir + "/mock.prop"
	mockDotEnv = mockDir + "/mock.env"
)

// Ordered list of fields in the 'flat' test confs.
var flatTestKeys = []string{
	"COMMENT",
	"MORE",
	"MiXeD",
	"UPPER",
	"empty",
	"lower",
	"quotedSpecial",
}

var flatTestKeyMap = map[string][]string{
	"COMMENT":       {"COMMENT"},
	"MORE":          {"MORE"},
	"MiXeD":         {"MiXeD"},
	"UPPER":         {"UPPER"},
	"empty":         {"empty"},
	"lower":         {"lower"},
	"quotedSpecial": {"quotedSpecial"},
}

var flatTestAll = `COMMENT -> AFTER
MORE -> vars
MiXeD -> CaSe
UPPER -> CASE
empty -> 
lower -> case
quotedSpecial -> j18120734xn2&*@#*&R#d1j23d*(*)`

// Ordered list of fields in the test confs.
var testKeys = []string{
	"bools",
	"duration",
	"empty",
	"intbools",
	"negative_int",
	"orphan",
	"parent1.boolmap.notok3",
	"parent1.boolmap.ok1",
	"parent1.boolmap.ok2",
	"parent1.child1.empty",
	"parent1.child1.grandchild1.ids",
	"parent1.child1.grandchild1.on",
	"parent1.child1.name",
	"parent1.child1.type",
	"parent1.floatmap.key1",
	"parent1.floatmap.key2",
	"parent1.floatmap.key3",
	"parent1.id",
	"parent1.intmap.key1",
	"parent1.intmap.key2",
	"parent1.intmap.key3",
	"parent1.name",
	"parent1.strmap.key1",
	"parent1.strmap.key2",
	"parent1.strmap.key3",
	"parent1.strsmap.key1",
	"parent1.strsmap.key2",
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
	"bools":                          {"bools"},
	"duration":                       {"duration"},
	"empty":                          {"empty"},
	"intbools":                       {"intbools"},
	"negative_int":                   {"negative_int"},
	"orphan":                         {"orphan"},
	"parent1":                        {"parent1"},
	"parent1.boolmap":                {"parent1", "boolmap"},
	"parent1.boolmap.notok3":         {"parent1", "boolmap", "notok3"},
	"parent1.boolmap.ok1":            {"parent1", "boolmap", "ok1"},
	"parent1.boolmap.ok2":            {"parent1", "boolmap", "ok2"},
	"parent1.child1":                 {"parent1", "child1"},
	"parent1.child1.empty":           {"parent1", "child1", "empty"},
	"parent1.child1.grandchild1":     {"parent1", "child1", "grandchild1"},
	"parent1.child1.grandchild1.ids": {"parent1", "child1", "grandchild1", "ids"},
	"parent1.child1.grandchild1.on":  {"parent1", "child1", "grandchild1", "on"},
	"parent1.child1.name":            {"parent1", "child1", "name"},
	"parent1.child1.type":            {"parent1", "child1", "type"},
	"parent1.floatmap":               {"parent1", "floatmap"},
	"parent1.floatmap.key1":          {"parent1", "floatmap", "key1"},
	"parent1.floatmap.key2":          {"parent1", "floatmap", "key2"},
	"parent1.floatmap.key3":          {"parent1", "floatmap", "key3"},
	"parent1.id":                     {"parent1", "id"},
	"parent1.intmap":                 {"parent1", "intmap"},
	"parent1.intmap.key1":            {"parent1", "intmap", "key1"},
	"parent1.intmap.key2":            {"parent1", "intmap", "key2"},
	"parent1.intmap.key3":            {"parent1", "intmap", "key3"},
	"parent1.name":                   {"parent1", "name"},
	"parent1.strmap":                 {"parent1", "strmap"},
	"parent1.strmap.key1":            {"parent1", "strmap", "key1"},
	"parent1.strmap.key2":            {"parent1", "strmap", "key2"},
	"parent1.strmap.key3":            {"parent1", "strmap", "key3"},
	"parent1.strsmap":                {"parent1", "strsmap"},
	"parent1.strsmap.key1":           {"parent1", "strsmap", "key1"},
	"parent1.strsmap.key2":           {"parent1", "strsmap", "key2"},
	"parent2":                        {"parent2"},
	"parent2.child2":                 {"parent2", "child2"},
	"parent2.child2.empty":           {"parent2", "child2", "empty"},
	"parent2.child2.grandchild2":     {"parent2", "child2", "grandchild2"},
	"parent2.child2.grandchild2.ids": {"parent2", "child2", "grandchild2", "ids"},
	"parent2.child2.grandchild2.on":  {"parent2", "child2", "grandchild2", "on"},
	"parent2.child2.name":            {"parent2", "child2", "name"},
	"parent2.id":                     {"parent2", "id"},
	"parent2.name":                   {"parent2", "name"},
	"strbool":                        {"strbool"},
	"strbools":                       {"strbools"},
	"time":                           {"time"},
	"type":                           {"type"},
}

// `parent1.child1.type` and `type` are excluded as their
// values vary between files.
var testAll = `bools -> [true false true]
duration -> 3s
empty -> map[]
intbools -> [1 0 1]
negative_int -> -1234
orphan -> [red blue orange]
parent1.boolmap.notok3 -> false
parent1.boolmap.ok1 -> true
parent1.boolmap.ok2 -> true
parent1.child1.empty -> map[]
parent1.child1.grandchild1.ids -> [1 2 3]
parent1.child1.grandchild1.on -> true
parent1.child1.name -> child1
parent1.floatmap.key1 -> 1.1
parent1.floatmap.key2 -> 1.2
parent1.floatmap.key3 -> 1.3
parent1.id -> 1234
parent1.intmap.key1 -> 1
parent1.intmap.key2 -> 1
parent1.intmap.key3 -> 1
parent1.name -> parent1
parent1.strmap.key1 -> val1
parent1.strmap.key2 -> val2
parent1.strmap.key3 -> val3
parent1.strsmap.key1 -> [val1 val2 val3]
parent1.strsmap.key2 -> [val4 val5]
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

type testStructFlat struct {
	Type                        string            `koanf:"type"`
	Empty                       map[string]string `koanf:"empty"`
	Parent1Name                 string            `koanf:"parent1.name"`
	Parent1ID                   int               `koanf:"parent1.id"`
	Parent1Child1Name           string            `koanf:"parent1.child1.name"`
	Parent1Child1Type           string            `koanf:"parent1.child1.type"`
	Parent1Child1Empty          map[string]string `koanf:"parent1.child1.empty"`
	Parent1Child1Grandchild1IDs []int             `koanf:"parent1.child1.grandchild1.ids"`
	Parent1Child1Grandchild1On  bool              `koanf:"parent1.child1.grandchild1.on"`
}

type Case struct {
	koanf    *koanf.Koanf
	file     string
	parser   koanf.Parser
	typeName string
}

// 'Flat' Case instances to be used in multiple tests. These will not be
// mutated.
var flatCases = []Case{
	{koanf: koanf.New(delim), file: mockDotEnv, parser: dotenv.Parser(), typeName: "dotenv"},
}

// Case instances to be used in multiple tests. These will not be mutated.
var cases = []Case{
	{koanf: koanf.New(delim), file: mockJSON, parser: json.Parser(), typeName: "json"},
	{koanf: koanf.New(delim), file: mockYAML, parser: yaml.Parser(), typeName: "yml"},
	{koanf: koanf.New(delim), file: mockTOML, parser: toml.Parser(), typeName: "toml"},
	{koanf: koanf.New(delim), file: mockHCL, parser: hcl.Parser(true), typeName: "hcl"},
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

	// Preload 1 'flat' Koanf instances with their providers and config.
	if err := flatCases[0].koanf.Load(file.Provider(flatCases[0].file), dotenv.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
}

func BenchmarkLoadFile(b *testing.B) {
	k := koanf.New(delim)

	// Don't use TOML here because it distorts memory benchmarks due to high memory use
	providers := []*file.File{file.Provider(mockJSON), file.Provider(mockYAML)}
	parsers := []koanf.Parser{json.Parser(), yaml.Parser()}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := k.Load(providers[n%2], parsers[n%2]); err != nil {
			b.Fatalf("Unexpected error: %+v", k)
		}
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

func TestLoadFlatFileAllKeys(t *testing.T) {
	assert := assert.New(t)
	re, _ := regexp.Compile("(.+?)?type \\-> (.*)\n")
	for _, c := range flatCases {
		// Check against testKeys.
		assert.Equal(flatTestKeys, c.koanf.Keys(), fmt.Sprintf("loaded keys mismatch: %v", c.typeName))

		// Check against keyMap.
		assert.EqualValues(flatTestKeyMap, c.koanf.KeyMap(), "keymap doesn't match")

		// Replace the "type" fields that varies across different files
		// to do a complete key -> value map match with testAll.
		s := strings.TrimSpace(re.ReplaceAllString(c.koanf.Sprint(), ""))
		assert.Equal(flatTestAll, s, fmt.Sprintf("key -> value list mismatch: %v", c.typeName))
	}
}

func TestLoadFileAllKeys(t *testing.T) {
	assert := assert.New(t)
	re, _ := regexp.Compile("(.+?)?type \\-> (.*)\n")
	for _, c := range cases {
		// Check against testKeys.
		assert.Equal(testKeys, c.koanf.Keys(), fmt.Sprintf("loaded keys mismatch: %v", c.typeName))

		// Check against keyMap.
		assert.EqualValues(testKeyMap, c.koanf.KeyMap(), "keymap doesn't match")

		// Replace the "type" fields that varies across different files
		// to do a complete key -> value map match with testAll.
		s := strings.TrimSpace(re.ReplaceAllString(c.koanf.Sprint(), ""))
		assert.Equal(testAll, s, fmt.Sprintf("key -> value list mismatch: %v", c.typeName))
	}
}

func TestDelim(t *testing.T) {
	k1 := koanf.New(".")
	assert.Equal(t, k1.Delim(), ".")

	k2 := koanf.New("/")
	assert.Equal(t, k2.Delim(), "/")
}

func TestLoadMergeYamlJson(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	assert.NoError(k.Load(file.Provider(mockYAML), yaml.Parser()),
		"error loading file")
	// loading json after yaml causes the intbools to be loaded as []float64
	assert.NoError(k.Load(file.Provider(mockJSON), yaml.Parser()),
		"error loading file")

	// checking that there is no issues with expecting it to be an []int64
	v := k.Int64s("intbools")
	assert.Len(v, 3)

	defer func() {
		if err := recover(); err != nil {
			assert.Failf("panic", "received panic: %v", err)
		}
	}()

	v2 := k.MustInt64s("intbools")
	assert.Len(v2, 3)
}

func TestLoadMergeJsonYaml(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	assert.NoError(k.Load(file.Provider(mockJSON), yaml.Parser()),
		"error loading file")
	// loading yaml after json causes the intbools to be loaded as []int after json loaded it with []float64
	assert.NoError(k.Load(file.Provider(mockYAML), yaml.Parser()),
		"error loading file")

	// checking that there is no issues with expecting it to be an []float64
	v := k.Float64s("intbools")
	assert.Len(v, 3)

	defer func() {
		if err := recover(); err != nil {
			assert.Failf("panic", "received panic: %v", err)
		}
	}()

	v2 := k.MustFloat64s("intbools")
	assert.Len(v2, 3)
}

func TestWatchFile(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	// Create a tmp config file.
	tmpDir, _ := ioutil.TempDir("", "koanf_*") // TODO: replace with t.TempDir() as of go v1.15
	tmpFile := filepath.Join(tmpDir, "koanf_mock")
	err := ioutil.WriteFile(tmpFile, []byte(`{"parent": {"name": "name1"}}`), 0600) // TODO: replace with os.WriteFile as of go v1.16
	require.NoError(t, err, "error creating temp config file: %v", err)

	// Load the new config and watch it for changes.
	f := file.Provider(tmpFile)
	k.Load(f, json.Parser())

	// Watch for changes.
	changedC := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1) // our assurance that cb is called max once
	f.Watch(func(event interface{}, err error) {
		// The File watcher always returns a nil `event`, which can
		// be ignored.
		if err != nil {
			// TODO: replace make with of Error Wrapping-Scheme and assert.ErrorIs() checks as of go v1.13
			assert.Condition(func() bool {
				return strings.Contains(err.Error(), "was removed")
			}, "received unexpected error. err: %s", err)
			return
		}
		// Reload the config.
		k.Load(f, json.Parser())
		changedC <- k.String("parent.name")
		wg.Done()
	})

	// Wait a second and change the file.
	time.Sleep(1 * time.Second)
	ioutil.WriteFile(tmpFile, []byte(`{"parent": {"name": "name2"}}`), 0600) // TODO: replace with os.WriteFile as of go v1.16
	wg.Wait()

	assert.Condition(func() bool {
		return strings.Compare(<-changedC, "name2") == 0
	}, "file watch reload didn't change config")
}

func TestWatchFileSymlink(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	tmpDir, _ := ioutil.TempDir("", "koanf_*") // TODO: replace with t.TempDir() as of go v1.15

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
	f.Watch(func(event interface{}, err error) {
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
	assert.NoError(os.Rename(symPath2, symPath), "error swaping symlink to another file type")
	wg.Wait()

	assert.Condition(func() bool {
		return strings.Compare(<-changedC, "yml") == 0
	}, "symlink watch reload didn't change config")
}

func TestLoadMerge(t *testing.T) {
	var (
		assert = assert.New(t)
		// Load several types into a fresh Koanf instance.
		k = koanf.New(delim)
	)
	for _, c := range cases {
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Check against testKeys.
		assert.Equal(testKeys, k.Keys(), fmt.Sprintf("loaded keys don't match in: %v", c.file))

		// The 'type' fields in different file types have different values.
		// As each subsequent file is loaded, the previous value should be overridden.
		assert.Equal(c.typeName, k.String("type"), "types don't match")
		assert.Equal(c.typeName, k.String("parent1.child1.type"), "types don't match")
	}

	// Load env provider and override value.
	os.Setenv("PREFIX_PARENT1.CHILD1.TYPE", "env")
	err := k.Load(env.Provider("PREFIX_", ".", func(k string) string {
		return strings.Replace(strings.ToLower(k), "prefix_", "", -1)
	}), nil)

	assert.Nil(err, "error loading env")
	assert.Equal("env", k.String("parent1.child1.type"), "types don't match")

	// Test the env provider than can mutate the value to upper case
	err = k.Load(env.ProviderWithValue("PREFIX_", ".", func(k string, v string) (string, interface{}) {
		return strings.Replace(strings.ToLower(k), "prefix_", "", -1), strings.ToUpper(v)
	}), nil)

	assert.Nil(err, "error loading env with value")
	assert.Equal("ENV", k.String("parent1.child1.type"), "types don't match")

	// Override with the confmap provider.
	k.Load(confmap.Provider(map[string]interface{}{
		"parent1.child1.type": "confmap",
		"type":                "confmap",
	}, "."), nil)
	assert.Equal("confmap", k.String("parent1.child1.type"), "types don't match")
	assert.Equal("confmap", k.String("type"), "types don't match")

	// Override with the rawbytes provider.
	assert.Nil(k.Load(rawbytes.Provider([]byte(`{"type": "rawbytes", "parent1": {"child1": {"type": "rawbytes"}}}`)), json.Parser()),
		"error loading raw bytes")
	assert.Equal("rawbytes", k.String("parent1.child1.type"), "types don't match")
	assert.Equal("rawbytes", k.String("type"), "types don't match")
}

func TestFlags(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	k2 := k.Copy()
	k3 := k.Copy()

	// Override with the posflag provider.
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)

	// Key that exists in the loaded conf. Should overwrite with Set().
	f.String("parent1.child1.type", "flag", "")
	f.Set("parent1.child1.type", "flag")
	f.StringSlice("stringslice", []string{"a", "b", "c"}, "")
	f.IntSlice("intslice", []int{1, 2, 3}, "")

	// Key that doesn't exist in the loaded file conf. Should merge the default value.
	f.String("flagkey", "flag", "")

	// Key that exists in the loadd conf but no Set(). Default value shouldn't be merged.
	f.String("parent1.name", "flag", "")

	// Initialize the provider with the Koanf instance passed where default values
	// will merge if the keys are not present in the conf map.
	assert.Nil(k.Load(posflag.Provider(f, ".", k), nil), "error loading posflag")
	assert.Equal("flag", k.String("parent1.child1.type"), "types don't match")
	assert.Equal("flag", k.String("flagkey"), "value doesn't match")
	assert.NotEqual("flag", k.String("parent1.name"), "value doesn't match")
	assert.Equal([]string{"a", "b", "c"}, k.Strings("stringslice"), "value doesn't match")
	assert.Equal([]int{1, 2, 3}, k.Ints("intslice"), "value doesn't match")

	// Test the posflag provider can mutate the value to upper case
	assert.Nil(k3.Load(posflag.ProviderWithValue(f, ".", nil, func(k string, v string) (string, interface{}) {
		return strings.Replace(strings.ToLower(k), "prefix_", "", -1), strings.ToUpper(v)
	}), nil), "error loading posflag")
	assert.Equal("FLAG", k3.String("parent1.child1.type"), "types don't match")

	// Test without passing the Koanf instance where default values will not merge.
	assert.Nil(k2.Load(posflag.Provider(f, ".", nil), nil), "error loading posflag")
	assert.Equal("flag", k2.String("parent1.child1.type"), "types don't match")
	assert.Equal("", k2.String("flagkey"), "value doesn't match")
	assert.NotEqual("", k2.String("parent1.name"), "value doesn't match")

	// Override with the flag provider.
	bf := flag.NewFlagSet("test", flag.ContinueOnError)
	bf.String("parent1.child1.type", "flag", "")
	bf.Set("parent1.child1.type", "basicflag")
	assert.Nil(k.Load(basicflag.Provider(bf, "."), nil), "error loading basicflag")
	assert.Equal("basicflag", k.String("parent1.child1.type"), "types don't match")

	// Test the basicflag provider can mutate the value to upper case
	bf2 := flag.NewFlagSet("test", flag.ContinueOnError)
	bf2.String("parent1.child1.type", "flag", "")
	bf2.Set("parent1.child1.type", "basicflag")
	assert.Nil(k.Load(basicflag.ProviderWithValue(bf2, ".", func(k string, v string) (string, interface{}) {
		return strings.Replace(strings.ToLower(k), "prefix_", "", -1), strings.ToUpper(v)
	}), nil), "error loading basicflag")
	assert.Equal("BASICFLAG", k.String("parent1.child1.type"), "types don't match")

}

func TestConfMapValues(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	var (
		c1  = k.All()
		ra1 = k.Raw()
		r1  = k.Cut("parent2").Raw()
	)

	k = koanf.New(delim)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	var (
		c2  = k.All()
		ra2 = k.Raw()
		r2  = k.Cut("parent2").Raw()
	)

	assert.EqualValues(c1, c2, "conf map mismatch")
	assert.EqualValues(ra1, ra2, "conf map mismatch")
	assert.EqualValues(r1, r2, "conf map mismatch")
}

func TestCutCopy(t *testing.T) {
	// Instance 1.
	var (
		assert = assert.New(t)
		k1     = koanf.New(delim)
	)
	assert.Nil(k1.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	var (
		cp1   = k1.Copy()
		cut1  = k1.Cut("")
		cutp1 = k1.Cut("parent2")
	)

	// Instance 2.
	k2 := koanf.New(delim)
	assert.Nil(k2.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	var (
		cp2   = k2.Copy()
		cut2  = k2.Cut("")
		cutp2 = k2.Cut("parent2")
	)

	assert.EqualValues(cp1.All(), cp2.All(), "conf map mismatch")
	assert.EqualValues(cut1.All(), cut2.All(), "conf map mismatch")
	assert.EqualValues(cutp1.All(), cutp2.All(), "conf map mismatch")
	assert.Equal(testParent2, strings.TrimSpace(cutp1.Sprint()), "conf map mismatch")
	assert.Equal(strings.TrimSpace(cutp1.Sprint()), strings.TrimSpace(cutp2.Sprint()), "conf map mismatch")

	// Cut a single field with no children. Should return empty conf maps.
	assert.Equal(k1.Cut("type").Keys(), k2.Cut("type").Keys(), "single field cut mismatch")
	assert.Equal(k1.Cut("xxxx").Keys(), k2.Cut("xxxx").Keys(), "single field cut mismatch")
	assert.Equal(len(k1.Cut("type").Raw()), 0, "non-map cut returned items")
}

func TestMerge(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")

	// Make two different cuts that'll have different confmaps.
	var (
		cut1 = k.Cut("parent1")
		cut2 = k.Cut("parent2")
	)
	assert.NotEqual(cut1.All(), cut2.All(), "different cuts incorrectly match")
	assert.NotEqual(cut1.Sprint(), cut2.Sprint(), "different cuts incorrectly match")

	// Create an empty Koanf instance.
	k2 := koanf.New(delim)

	// Merge cut1 into it and check if they match.
	k2.Merge(cut1)
	assert.Equal(cut1.All(), k2.All(), "conf map mismatch")
}

func TestRaw_YamlTypes(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	assert.Nil(k.Load(file.Provider(mockYAML), yaml.Parser()),
		"error loading file")
	raw := k.Raw()

	i, ok := raw["intbools"]
	assert.True(ok, "ints key does not exist in the map")

	arr, ok := i.([]interface{})
	assert.True(ok, "arr slice is not array of integers")

	for _, integer := range arr {
		if _, ok := integer.(int); !ok {
			assert.Failf("failure", "%v not an integer but %T", integer, integer)
		}
	}
}

func TestRaw_JsonTypes(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	raw := k.Raw()

	i, ok := raw["intbools"]
	assert.True(ok, "ints key does not exist in the map")

	arr, ok := i.([]interface{})
	assert.True(ok, "arr slice is not array of integers")

	for _, integer := range arr {
		if _, ok := integer.(float64); !ok {
			assert.Failf("failure", "%v not an integer but %T", integer, integer)
		}
	}
}

func TestMergeStrictError(t *testing.T) {
	var (
		assert = assert.New(t)
	)

	ks := koanf.NewWithConf(koanf.Conf{
		Delim:       delim,
		StrictMerge: true,
	})

	assert.Nil(ks.Load(confmap.Provider(map[string]interface{}{
		"parent2": map[string]interface{}{
			"child2": map[string]interface{}{
				"grandchild2": map[string]interface{}{
					"ids": 123,
				},
			},
		},
	}, delim), nil))

	err := ks.Load(file.Provider(mockYAML), yaml.Parser())
	assert.Error(err)
	assert.True(strings.HasPrefix(err.Error(), "incorrect types at key parent2.child2.grandchild2"))
}

func TestMergeAt(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockYAML), yaml.Parser()),
		"error loading file")

	// Get expected koanf, and root data
	var (
		expected = k.Cut("parent2")
		rootData = map[string]interface{}{
			"name": k.String("parent2.name"),
			"id":   k.Int("parent2.id"),
		}
	)

	// Get nested test data to merge at path
	child2 := koanf.New(delim)
	assert.Nil(child2.Load(confmap.Provider(map[string]interface{}{
		"name":  k.String("parent2.child2.name"),
		"empty": k.Get("parent2.child2.empty"),
	}, delim), nil))
	grandChild := k.Cut("parent2.child2.grandchild2")

	// Create test koanf
	ordered := koanf.New(delim)
	assert.Nil(ordered.Load(confmap.Provider(rootData, delim), nil))

	// Merge at path in order, first child2, then child2.grandchild2
	ordered.MergeAt(child2, "child2")
	ordered.MergeAt(grandChild, "child2.grandchild2")
	assert.Equal(expected.Get(""), ordered.Get(""), "conf map mismatch")

	// Create test koanf
	reversed := koanf.New(delim)
	assert.Nil(reversed.Load(confmap.Provider(rootData, delim), nil))

	// Merge at path in reverse order, first child2.grandchild2, then child2
	reversed.MergeAt(grandChild, "child2.grandchild2")
	reversed.MergeAt(child2, "child2")
	assert.Equal(expected.Get(""), reversed.Get(""), "conf map mismatch")
}

func TestUnmarshal(t *testing.T) {
	assert := assert.New(t)

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
			k  = koanf.New(delim)
			ts testStruct
		)
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		assert.Nil(k.Unmarshal("", &ts), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")

		// Unmarshal with config.
		ts = testStruct{}
		assert.Nil(k.UnmarshalWithConf("", &ts, koanf.UnmarshalConf{Tag: "koanf"}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")
	}
}

func TestUnmarshalFlat(t *testing.T) {
	assert := assert.New(t)

	// Expected unmarshalled structure.
	real := testStructFlat{
		Type:                        "json",
		Empty:                       make(map[string]string),
		Parent1Name:                 "parent1",
		Parent1ID:                   1234,
		Parent1Child1Name:           "child1",
		Parent1Child1Type:           "json",
		Parent1Child1Empty:          make(map[string]string),
		Parent1Child1Grandchild1IDs: []int{1, 2, 3},
		Parent1Child1Grandchild1On:  true,
	}

	// Unmarshal and check all parsers.
	for _, c := range cases {
		k := koanf.New(delim)
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		ts := testStructFlat{}
		assert.Nil(k.UnmarshalWithConf("", &ts, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1Child1Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")
	}
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	for _, c := range cases {
		// HCL does not support marshalling.
		if c.typeName == "hcl" {
			continue
		}

		// Load config.
		k := koanf.New(delim)
		assert.NoError(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Serialize / marshal into raw bytes using the parser.
		b, err := k.Marshal(c.parser)
		assert.NoError(err, "error marshalling")

		// Reload raw serialize bytes into a new koanf instance.
		k = koanf.New(delim)
		assert.NoError(k.Load(rawbytes.Provider(b), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Check if values are intact.
		assert.Equal(float64(1234), c.koanf.MustFloat64("parent1.id"))
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.MustStrings("orphan"))
		assert.Equal([]int64{1, 2, 3}, c.koanf.MustInt64s("parent1.child1.grandchild1.ids"))
	}
}

func TestGetExists(t *testing.T) {
	assert := assert.New(t)

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
		assert.Equal(c.exists, cases[0].koanf.Exists(c.path),
			fmt.Sprintf("path resolution failed: %s", c.path))
		assert.Equal(c.exists, cases[0].koanf.Get(c.path) != nil,
			fmt.Sprintf("path resolution failed: %s", c.path))
	}
}

func TestSlices(t *testing.T) {
	assert := assert.New(t)

	// Load a slice of confmaps [{}, {}].
	var mp map[string]interface{}
	err := encjson.Unmarshal([]byte(`{
		"parent": [
			{"value": 1, "sub": {"value": "1"}},
			{"value": 2, "sub": {"value": "2"}}
		],
		"another": "123"
	}`), &mp)
	assert.NoError(err, "error marshalling test payload")
	k := koanf.New(delim)
	assert.NoError(k.Load(confmap.Provider(mp, "."), nil))

	assert.Empty(k.Slices("x"))
	assert.Empty(k.Slices("parent.value"))
	assert.Empty(k.Slices("parent.value.sub"))

	slices := k.Slices("parent")
	assert.NotNil(slices, "got nil slice of confmap")
	assert.NotEmpty(slices, "got empty confmap slice")

	for i, s := range slices {
		assert.Equal(s.Int("value"), i+1)
		assert.Equal(s.String("sub.value"), fmt.Sprintf("%d", i+1))
	}
}

func TestGetTypes(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		assert.Equal(nil, c.koanf.Get("xxx"))
		assert.Equal(make(map[string]interface{}), c.koanf.Get("empty"))

		// Int.
		assert.Equal(int64(0), c.koanf.Int64("xxxx"))
		assert.Equal(int64(1234), c.koanf.Int64("parent1.id"))

		assert.Equal(int(0), c.koanf.Int("xxxx"))
		assert.Equal(int(1234), c.koanf.Int("parent1.id"))

		assert.Equal([]int64{}, c.koanf.Int64s("xxxx"))
		assert.Equal([]int64{1, 2, 3}, c.koanf.Int64s("parent1.child1.grandchild1.ids"))

		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Int64Map("parent1.intmap"))
		assert.Equal(map[string]int64{}, c.koanf.Int64Map("parent1.boolmap"))
		assert.Equal(map[string]int64{}, c.koanf.Int64Map("xxxx"))
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Int64Map("parent1.floatmap"))

		assert.Equal([]int{1, 2, 3}, c.koanf.Ints("parent1.child1.grandchild1.ids"))
		assert.Equal([]int{}, c.koanf.Ints("xxxx"))

		assert.Equal(map[string]int{"key1": 1, "key2": 1, "key3": 1}, c.koanf.IntMap("parent1.intmap"))
		assert.Equal(map[string]int{}, c.koanf.IntMap("parent1.boolmap"))
		assert.Equal(map[string]int{}, c.koanf.IntMap("xxxx"))

		// Float.
		assert.Equal(float64(0), c.koanf.Float64("xxx"))
		assert.Equal(float64(1234), c.koanf.Float64("parent1.id"))

		assert.Equal([]float64{}, c.koanf.Float64s("xxxx"))
		assert.Equal([]float64{1, 2, 3}, c.koanf.Float64s("parent1.child1.grandchild1.ids"))

		assert.Equal(map[string]float64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Float64Map("parent1.intmap"))
		assert.Equal(map[string]float64{"key1": 1.1, "key2": 1.2, "key3": 1.3}, c.koanf.Float64Map("parent1.floatmap"))
		assert.Equal(map[string]float64{}, c.koanf.Float64Map("parent1.boolmap"))
		assert.Equal(map[string]float64{}, c.koanf.Float64Map("xxxx"))

		// String and bytes.
		assert.Equal([]byte{}, c.koanf.Bytes("xxxx"))
		assert.Equal([]byte("parent1"), c.koanf.Bytes("parent1.name"))

		assert.Equal("", c.koanf.String("xxxx"))
		assert.Equal("parent1", c.koanf.String("parent1.name"))

		assert.Equal([]string{}, c.koanf.Strings("xxxx"))
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.Strings("orphan"))

		assert.Equal(map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}, c.koanf.StringMap("parent1.strmap"))
		assert.Equal(map[string][]string{"key1": {"val1", "val2", "val3"}, "key2": {"val4", "val5"}}, c.koanf.StringsMap("parent1.strsmap"))
		assert.Equal(map[string]string{}, c.koanf.StringMap("xxxx"))
		assert.Equal(map[string]string{}, c.koanf.StringMap("parent1.intmap"))

		// Bools.
		assert.Equal(false, c.koanf.Bool("xxxx"))
		assert.Equal(false, c.koanf.Bool("type"))
		assert.Equal(true, c.koanf.Bool("parent1.child1.grandchild1.on"))
		assert.Equal(true, c.koanf.Bool("strbool"))

		assert.Equal([]bool{}, c.koanf.Bools("xxxx"))
		assert.Equal([]bool{true, false, true}, c.koanf.Bools("bools"))
		assert.Equal([]bool{true, false, true}, c.koanf.Bools("intbools"))
		assert.Equal([]bool{true, true, false}, c.koanf.Bools("strbools"))

		assert.Equal(map[string]bool{"ok1": true, "ok2": true, "notok3": false}, c.koanf.BoolMap("parent1.boolmap"))
		assert.Equal(map[string]bool{"key1": true, "key2": true, "key3": true}, c.koanf.BoolMap("parent1.intmap"))
		assert.Equal(map[string]bool{}, c.koanf.BoolMap("xxxx"))

		// Others.
		assert.Equal(time.Duration(1234), c.koanf.Duration("parent1.id"))
		assert.Equal(time.Duration(0), c.koanf.Duration("xxxx"))
		assert.Equal(time.Second*3, c.koanf.Duration("duration"))

		assert.Equal(time.Time{}, c.koanf.Time("xxxx", "2006-01-02"))
		assert.Equal(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), c.koanf.Time("time", "2006-01-02"))

		assert.Equal([]string{}, c.koanf.MapKeys("xxxx"), "map keys mismatch")
		assert.Equal([]string{"bools", "duration", "empty", "intbools", "negative_int", "orphan", "parent1", "parent2", "strbool", "strbools", "time", "type"},
			c.koanf.MapKeys(""), "map keys mismatch")
		assert.Equal([]string{"key1", "key2", "key3"}, c.koanf.MapKeys("parent1.strmap"), "map keys mismatch")

		// Attempt to parse int=1234 as a Unix timestamp.
		assert.Equal(time.Date(1970, 1, 1, 0, 20, 34, 0, time.UTC), c.koanf.Time("parent1.id", "").UTC())
	}
}

func TestMustGetTypes(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		// Int.
		assert.Panics(func() { c.koanf.MustInt64("xxxx") })
		assert.Equal(int64(1234), c.koanf.MustInt64("parent1.id"))
		assert.Equal(int64(-1234), c.koanf.MustInt64("negative_int"))

		assert.Panics(func() { c.koanf.MustInt("xxxx") })
		assert.Equal(int(1234), c.koanf.MustInt("parent1.id"))
		assert.Equal(int(-1234), c.koanf.MustInt("negative_int"))

		assert.Panics(func() { c.koanf.MustInt64s("xxxx") })
		assert.Equal([]int64{1, 2, 3}, c.koanf.MustInt64s("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustInt64Map("xxxx") })
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustInt64Map("parent1.intmap"))

		assert.Panics(func() { c.koanf.MustInt64Map("parent1.boolmap") })
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustInt64Map("parent1.floatmap"))

		assert.Panics(func() { c.koanf.MustInts("xxxx") })
		assert.Equal([]int{1, 2, 3}, c.koanf.MustInts("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustIntMap("xxxx") })
		assert.Panics(func() { c.koanf.MustIntMap("parent1.boolmap") })
		assert.Equal(map[string]int{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustIntMap("parent1.intmap"))

		// Float.
		assert.Panics(func() { c.koanf.MustInts("xxxx") })
		assert.Equal(float64(1234), c.koanf.MustFloat64("parent1.id"))

		assert.Panics(func() { c.koanf.MustFloat64s("xxxx") })
		assert.Equal([]float64{1, 2, 3}, c.koanf.MustFloat64s("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustFloat64Map("xxxx") })
		assert.Panics(func() { c.koanf.MustFloat64Map("parent1.boolmap") })
		assert.Equal(map[string]float64{"key1": 1.1, "key2": 1.2, "key3": 1.3}, c.koanf.MustFloat64Map("parent1.floatmap"))
		assert.Equal(map[string]float64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustFloat64Map("parent1.intmap"))

		// String and bytes.
		assert.Panics(func() { c.koanf.MustBytes("xxxx") })
		assert.Equal([]byte("parent1"), c.koanf.MustBytes("parent1.name"))

		assert.Panics(func() { c.koanf.MustString("xxxx") })
		assert.Equal("parent1", c.koanf.MustString("parent1.name"))

		assert.Panics(func() { c.koanf.MustStrings("xxxx") })
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.MustStrings("orphan"))

		assert.Panics(func() { c.koanf.MustStringMap("xxxx") })
		assert.Panics(func() { c.koanf.MustStringMap("parent1.intmap") })
		assert.Equal(map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}, c.koanf.MustStringMap("parent1.strmap"))
		assert.Equal(map[string][]string{"key1": {"val1", "val2", "val3"}, "key2": {"val4", "val5"}}, c.koanf.MustStringsMap("parent1.strsmap"))

		// // Bools.
		assert.Panics(func() { c.koanf.MustBools("xxxx") })
		assert.Equal([]bool{true, false, true}, c.koanf.MustBools("bools"))
		assert.Equal([]bool{true, false, true}, c.koanf.MustBools("intbools"))
		assert.Equal([]bool{true, true, false}, c.koanf.MustBools("strbools"))

		assert.Panics(func() { c.koanf.MustBoolMap("xxxx") })
		assert.Equal(map[string]bool{"ok1": true, "ok2": true, "notok3": false}, c.koanf.MustBoolMap("parent1.boolmap"))
		assert.Equal(map[string]bool{"key1": true, "key2": true, "key3": true}, c.koanf.MustBoolMap("parent1.intmap"))

		// Others.
		assert.Panics(func() { c.koanf.MustDuration("xxxx") })
		assert.Equal(time.Duration(1234), c.koanf.MustDuration("parent1.id"))
		assert.Equal(time.Second*3, c.koanf.MustDuration("duration"))
		assert.Equal(time.Duration(-1234), c.koanf.MustDuration("negative_int"))

		assert.Panics(func() { c.koanf.MustTime("xxxx", "2006-01-02") })
		assert.Equal(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), c.koanf.MustTime("time", "2006-01-02"))

		// // Attempt to parse int=1234 as a Unix timestamp.
		assert.Panics(func() { c.koanf.MustTime("time", "2006") })
		assert.Equal(time.Date(1970, 1, 1, 0, 20, 34, 0, time.UTC), c.koanf.MustTime("parent1.id", "").UTC())
	}
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		c.koanf.Delete("parent2.child2.grandchild2")
		assert.Equal(false, c.koanf.Exists("parent2.child2.grandchild2.on"))
		assert.Equal(false, c.koanf.Exists("parent2.child2.grandchild2.ids.5"))
		assert.Equal(true, c.koanf.Exists("parent2.child2.name"))

		c.koanf.Delete("")
		assert.Equal(false, c.koanf.Exists("duration"))
		assert.Equal(false, c.koanf.Exists("empty"))
	}
}
