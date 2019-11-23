package koanf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf/maps"
	"github.com/mitchellh/mapstructure"
)

// Koanf is the configuration apparatus.
type Koanf struct {
	confMap     map[string]interface{}
	confMapFlat map[string]interface{}
	keyMap      KeyMap
	delim       string
}

// KeyMap represents a map of flattened delimited keys and the non-delimited
// parts as their slices. For nested keys, the map holds all levels of path combinations.
// For example, the nested structure `parent -> child -> key` will produce the map:
// parent.child.key => [parent, child, key]
// parent.child => [parent, child]
// parent => [parent]
type KeyMap map[string][]string

// UnmarshalConf represents configuration options used by
// Unmarshal() to unmarshal conf maps into arbitrary structs.
type UnmarshalConf struct {
	// Tag is the struct field tag to unmarshal.
	// `koanf` is used if left empty.
	Tag string

	// If this is set to true, instead of unmarshalling nested structures
	// based on the key path, keys are taken literally to unmarshal into
	// a flat struct. For example:
	// ```
	// type MyStuff struct {
	// 	Child1Name string `koanf:"parent1.child1.name"`
	// 	Child2Name string `koanf:"parent2.child2.name"`
	// 	Type       string `koanf:"json"`
	// }
	// ```
	FlatPaths     bool
	DecoderConfig *mapstructure.DecoderConfig
}

// New returns a new instance of Koanf. delim is the delimiter to use
// when specifying config key paths, for instance a . for `parent.child.key`
// or a / for `parent/child/key`.
func New(delim string) *Koanf {
	return &Koanf{
		delim:       delim,
		confMap:     make(map[string]interface{}),
		confMapFlat: make(map[string]interface{}),
		keyMap:      make(KeyMap),
	}
}

// Load takes a Provider that either provides a parsed config map[string]interface{}
// in which case pa (Parser) can be nil, or raw bytes to be parsed, where a Parser
// can be provided to parse.
func (ko *Koanf) Load(p Provider, pa Parser) error {
	var (
		mp  map[string]interface{}
		err error
	)

	// No Parser is given. Call the Provider's Read() methid to get
	// the config map.
	if pa == nil {
		mp, err = p.Read()
		if err != nil {
			return err
		}
	} else {
		// There's a Parser. Get raw bytes from the Provider to parse.
		b, err := p.ReadBytes()
		if err != nil {
			return err
		}
		mp, err = pa.Parse(b)
		if err != nil {
			return err
		}
	}

	ko.merge(mp)
	return nil
}

// Keys returns the slice of all flattened keys in the loaded configuration
// sorted alphabetically.
func (ko *Koanf) Keys() []string {
	out := make([]string, 0, len(ko.confMapFlat))
	for k := range ko.confMapFlat {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// KeyMap returns a map of flattened keys and the individual parts of the
// key as slices. eg: "parent.child.key" => ["parent", "child", "key"]
func (ko *Koanf) KeyMap() KeyMap {
	out := make(KeyMap, len(ko.keyMap))
	for key, parts := range ko.keyMap {
		out[key] = make([]string, len(parts))
		copy(out[key][:], parts[:])
	}
	return out
}

// All returns a map of all flattened key paths and their values.
// Note that it uses maps.Copy to create a copy that uses
// json.Marshal which changes the numeric types to float64.
func (ko *Koanf) All() map[string]interface{} {
	return maps.Copy(ko.confMapFlat)
}

// Raw returns a copy of the full raw conf map.
// Note that it uses maps.Copy to create a copy that uses
// json.Marshal which changes the numeric types to float64.
func (ko *Koanf) Raw() map[string]interface{} {
	return maps.Copy(ko.confMap)
}

// Sprint returns a key -> value string representation
// of the config map with keys sorted alphabetically.
func (ko *Koanf) Sprint() string {
	b := bytes.Buffer{}
	for _, k := range ko.Keys() {
		b.Write([]byte(fmt.Sprintf("%s -> %v\n", k, ko.confMapFlat[k])))
	}
	return b.String()
}

// Print prints a key -> value string representation
// of the config map with keys sorted alphabetically.
func (ko *Koanf) Print() {
	fmt.Print(ko.Sprint())
}

// Cut cuts the config map at a given key path into a sub map and
// returns a new Koanf instance with the cut config map loaded.
// For instance, if the loaded config has a path that looks like
// parent.child.sub.a.b, `Cut("parent.child")` returns a new Koanf
// instance with the config map `sub.a.b` where everything above
// `parent.child` are cut out.
func (ko *Koanf) Cut(path string) *Koanf {
	out := make(map[string]interface{})

	// Cut only makes sense if the requested key path is a map.
	if v, ok := ko.Get(path).(map[string]interface{}); ok {
		out = v
	}

	n := New(ko.delim)
	n.merge(out)
	return n
}

// Copy returns a copy of the Koanf instance.
func (ko *Koanf) Copy() *Koanf {
	return ko.Cut("")
}

// Merge merges the config map of a given Koanf instance into
// the current instance.
func (ko *Koanf) Merge(in *Koanf) {
	ko.merge(in.Raw())
}

// Unmarshal unmarshals a given key path into the given struct using
// the mapstructure lib. If no path is specified, the whole map is unmarshalled.
// `koanf` is the struct field tag used to match field names. To customize,
// use UnmarshalWithConf(). It uses the mitchellh/mapstructure package.
func (ko *Koanf) Unmarshal(path string, o interface{}) error {
	return ko.UnmarshalWithConf(path, o, UnmarshalConf{})
}

// UnmarshalWithConf is like Unmarshal but takes configuration params in UnmarshalConf.
// See mitchellh/mapstructure's DecoderConfig for advanced customization
// of the unmarshal behaviour.
func (ko *Koanf) UnmarshalWithConf(path string, o interface{}, c UnmarshalConf) error {
	if c.DecoderConfig == nil {
		c.DecoderConfig = &mapstructure.DecoderConfig{
			Metadata:         nil,
			Result:           o,
			WeaklyTypedInput: true,
		}
	}

	if c.Tag == "" {
		c.DecoderConfig.TagName = "koanf"
	} else {
		c.DecoderConfig.TagName = c.Tag
	}

	d, err := mapstructure.NewDecoder(c.DecoderConfig)
	if err != nil {
		return err
	}

	// Unmarshal using flat key paths.
	mp := ko.Get(path)
	if c.FlatPaths {
		if f, ok := mp.(map[string]interface{}); ok {
			fmp, _ := maps.Flatten(f, nil, ko.delim)
			mp = fmp
		}
	}

	return d.Decode(mp)
}

// Get returns the raw, uncast interface{} value of a given key path
// in the config map. If the key path does not exist, nil is returned.
func (ko *Koanf) Get(path string) interface{} {
	// No path. Return the whole conf map.
	if path == "" {
		return ko.Raw()
	}

	// Does the path exist?
	p, ok := ko.keyMap[path]
	if !ok {
		return nil
	}
	res := maps.Search(ko.confMap, p)

	// Non-reference types are okay to return directly.
	// Other types are "copied" with maps.Copy or json.Marshal
	// that change the numeric types to float64.

	switch v := res.(type) {
	case int, int8, int16, int32, int64, float32, float64, string, bool:
		return v
	case map[string]interface{}:
		return maps.Copy(v)
	}

	// Inefficient, but marshal and unmarshal to create a copy
	// of reference types to not expose  internal references to slices and maps.
	var out interface{}
	b, _ := json.Marshal(res)
	json.Unmarshal(b, &out)
	return out
}

// Exists returns true if the given key path exists in the conf map.
func (ko *Koanf) Exists(path string) bool {
	_, ok := ko.keyMap[path]
	return ok
}

// Int64 returns the int64 value of a given key path or 0 if the path
// does not exist or if the value is not a valid int64.
func (ko *Koanf) Int64(path string) int64 {
	if v := ko.Get(path); v != nil {
		i, _ := toInt64(v)
		return i
	}
	return 0
}

// Int64s returns the []int64 slice value of a given key path or an
// empty []int64 slice if the path does not exist or if the value
// is not a valid int slice.
func (ko *Koanf) Int64s(path string) []int64 {
	o := ko.Get(path)
	if o == nil {
		return []int64{}
	}

	var out []int64
	switch v := o.(type) {
	case []interface{}:
		out = make([]int64, 0, len(v))
		for _, vi := range v {
			i, err := toInt64(vi)

			// On error, return as it's not a valid
			// int slice.
			if err != nil {
				return []int64{}
			}
			out = append(out, i)
		}
		return out
	}

	return []int64{}
}

// Int64Map returns the map[string]int64 value of a given key path
// or an empty map[string]int64 if the path does not exist or if the
// value is not a valid int64 map.
func (ko *Koanf) Int64Map(path string) map[string]int64 {
	var (
		out = map[string]int64{}
		o   = ko.Get(path)
	)
	if o == nil {
		return out
	}

	mp, ok := o.(map[string]interface{})
	if !ok {
		return out
	}

	out = make(map[string]int64, len(mp))
	for k, v := range mp {
		switch i := v.(type) {
		case int64:
			out[k] = i
		default:
			// Attempt a conversion.
			iv, err := toInt64(i)
			if err != nil {
				return map[string]int64{}
			}
			out[k] = iv
		}
	}
	return out
}

// Int returns the int value of a given key path or 0 if the path
// does not exist or if the value is not a valid int.
func (ko *Koanf) Int(path string) int {
	return int(ko.Int64(path))
}

// Ints returns the []int slice value of a given key path or an
// empty []int slice if the path does not exist or if the value
// is not a valid int slice.
func (ko *Koanf) Ints(path string) []int {
	ints := ko.Int64s(path)
	if len(ints) == 0 {
		return []int{}
	}

	out := make([]int, len(ints))
	for i, v := range ints {
		out[i] = int(v)
	}
	return out
}

// IntMap returns the map[string]int value of a given key path
// or an empty map[string]int if the path does not exist or if the
// value is not a valid int map.
func (ko *Koanf) IntMap(path string) map[string]int {
	var (
		mp  = ko.Int64Map(path)
		out = make(map[string]int, len(mp))
	)
	for k, v := range mp {
		out[k] = int(v)
	}
	return out
}

// Float64 returns the float64 value of a given key path or 0 if the path
// does not exist or if the value is not a valid float64.
func (ko *Koanf) Float64(path string) float64 {
	if v := ko.Get(path); v != nil {
		f, _ := toFloat64(v)
		return f
	}
	return 0.0
}

// Float64s returns the []float64 slice value of a given key path or an
// empty []float64 slice if the path does not exist or if the value
// is not a valid float64 slice.
func (ko *Koanf) Float64s(path string) []float64 {
	o := ko.Get(path)
	if o == nil {
		return []float64{}
	}

	var out []float64
	switch v := o.(type) {
	case []interface{}:
		out = make([]float64, 0, len(v))
		for _, vi := range v {
			i, err := toFloat64(vi)

			// On error, return as it's not a valid
			// int slice.
			if err != nil {
				return []float64{}
			}
			out = append(out, float64(i))
		}
		return out
	}

	return []float64{}
}

// Float64Map returns the map[string]float64 value of a given key path
// or an empty map[string]float64 if the path does not exist or if the
// value is not a valid float64 map.
func (ko *Koanf) Float64Map(path string) map[string]float64 {
	var (
		out = map[string]float64{}
		o   = ko.Get(path)
	)
	if o == nil {
		return out
	}

	mp, ok := o.(map[string]interface{})
	if !ok {
		return out
	}

	out = make(map[string]float64, len(mp))
	for k, v := range mp {
		switch i := v.(type) {
		case float64:
			out[k] = i
		default:
			// Attempt a conversion.
			iv, err := toFloat64(i)
			if err != nil {
				return map[string]float64{}
			}
			out[k] = iv
		}
	}
	return out
}

// Duration returns the time.Duration value of a given key path assuming
// that the key contains a valid numeric value.
func (ko *Koanf) Duration(path string) time.Duration {
	return time.Duration(ko.Int64(path))
}

// Time attempts to parse the value of a given key path and return time.Time
// representation. If the value is numeric, it is treated as a UNIX timestamp
// and if it's string, a parse is attempted with the given layout.
func (ko *Koanf) Time(path, layout string) time.Time {
	// Unix timestamp?
	v := ko.Int64(path)
	if v != 0 {
		return time.Unix(v, 0)
	}

	// String representation.
	s := ko.String(path)
	if s != "" {
		t, _ := time.Parse(layout, s)
		return t
	}

	return time.Time{}
}

// String returns the string value of a given key path or "" if the path
// does not exist or if the value is not a valid string.
func (ko *Koanf) String(path string) string {
	if v := ko.Get(path); v != nil {
		if i, ok := v.(string); ok {
			return i
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// Strings returns the []string slice value of a given key path or an
// empty []string slice if the path does not exist or if the value
// is not a valid string slice.
func (ko *Koanf) Strings(path string) []string {
	o := ko.Get(path)
	if o == nil {
		return []string{}
	}

	var out []string
	switch v := o.(type) {
	case []interface{}:
		out = make([]string, 0, len(v))
		for _, u := range v {
			if s, ok := u.(string); ok {
				out = append(out, s)
			} else {
				out = append(out, fmt.Sprintf("%v", u))
			}
		}
		return out
	case []string:
		out := make([]string, len(v))
		copy(out[:], v[:])
		return out
	}
	return []string{}
}

// StringMap returns the map[string]string value of a given key path
// or an empty map[string]string if the path does not exist or if the
// value is not a valid string map.
func (ko *Koanf) StringMap(path string) map[string]string {
	var (
		out = map[string]string{}
		o   = ko.Get(path)
	)
	if o == nil {
		return out
	}

	mp, ok := o.(map[string]interface{})
	if !ok {
		return out
	}
	out = make(map[string]string, len(mp))
	for k, v := range mp {
		switch s := v.(type) {
		case string:
			out[k] = s
		default:
			// There's a non string type. Return.
			return map[string]string{}
		}
	}

	return out
}

// Bytes returns the []byte value of a given key path or an empty
// []byte slice if the path does not exist or if the value is not a valid string.
func (ko *Koanf) Bytes(path string) []byte {
	return []byte(ko.String(path))
}

// Bool returns the bool value of a given key path or false if the path
// does not exist or if the value is not a valid bool representation.
// Accepted string representations of bool are the ones supported by strconv.ParseBool.
func (ko *Koanf) Bool(path string) bool {
	if v := ko.Get(path); v != nil {
		b, _ := toBool(v)
		return b
	}
	return false
}

// Bools returns the []bool slice value of a given key path or an
// empty []bool slice if the path does not exist or if the value
// is not a valid bool slice.
func (ko *Koanf) Bools(path string) []bool {
	o := ko.Get(path)
	if o == nil {
		return []bool{}
	}

	var out []bool
	switch v := o.(type) {
	case []interface{}:
		out = make([]bool, 0, len(v))
		for _, u := range v {
			b, err := toBool(u)
			if err != nil {
				return nil
			}
			out = append(out, b)
		}
		return out
	case []bool:
		return out
	}
	return nil
}

// BoolMap returns the map[string]bool value of a given key path
// or an empty map[string]bool if the path does not exist or if the
// value is not a valid bool map.
func (ko *Koanf) BoolMap(path string) map[string]bool {
	var (
		out = map[string]bool{}
		o   = ko.Get(path)
	)
	if o == nil {
		return out
	}

	mp, ok := o.(map[string]interface{})
	if !ok {
		return out
	}
	out = make(map[string]bool, len(mp))
	for k, v := range mp {
		switch i := v.(type) {
		case bool:
			out[k] = i
		default:
			// Attempt a conversion.
			b, err := toBool(i)
			if err != nil {
				return map[string]bool{}
			}
			out[k] = b
		}
	}

	return out
}

// MapKeys returns a sorted string list of keys in a map addressed by the
// given path. If the path is not a map, an empty string slice is
// returned.
func (ko *Koanf) MapKeys(path string) []string {
	var (
		out = []string{}
		o   = ko.Get(path)
	)
	if o == nil {
		return out
	}

	mp, ok := o.(map[string]interface{})
	if !ok {
		return out
	}
	out = make([]string, 0, len(mp))
	for k := range mp {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (ko *Koanf) merge(c map[string]interface{}) {
	maps.IntfaceKeysToStrings(c)
	maps.Merge(c, ko.confMap)

	// Maintain a flattened version as well.
	ko.confMapFlat, ko.keyMap = maps.Flatten(ko.confMap, nil, ko.delim)
	ko.keyMap = populateKeyParts(ko.keyMap, ko.delim)
}

// toInt64 takes an interface value and if it is an integer type,
// converts and returns int64. If it's any other type,
// forces it to a string and attempts to an strconv.Atoi
// to get an integer out.
func toInt64(v interface{}) (int64, error) {
	switch i := v.(type) {
	case int:
		return int64(i), nil
	case int8:
		return int64(i), nil
	case int16:
		return int64(i), nil
	case int32:
		return int64(i), nil
	case int64:
		return i, nil
	}

	// Force it to a string and try to convert.
	f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	if err != nil {
		return 0, err
	}

	return int64(f), nil
}

// toInt64 takes an interface v interface{}value and if it is a float type,
// converts and returns float6v interface{}4. If it's any other type,
// forces it to a string and av interface{}ttempts to an strconv.ParseFloat
// to get a float out.
func toFloat64(v interface{}) (float64, error) {
	switch i := v.(type) {
	case float32:
		return float64(i), nil
	case float64:
		return i, nil
	}

	// Force it to a string and try to convert.
	f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	if err != nil {
		return f, err
	}

	return f, nil
}

// toBool takes an interface value and if it is a bool type,
// returns it. If it's any other type, forces it to a string and attempts
// to parse it as a bool using strconv.ParseBool.
func toBool(v interface{}) (bool, error) {
	if b, ok := v.(bool); ok {
		return b, nil
	}

	// Force it to a string and try to convert.
	b, err := strconv.ParseBool(fmt.Sprintf("%v", v))
	if err != nil {
		return b, err
	}
	return b, nil
}

// populateKeyParts iterates a key map and generates all possible
// traveral paths. For instance, `parent.child.key` generates
// `parent`, and `parent.child`.
func populateKeyParts(m KeyMap, delim string) KeyMap {
	out := make(KeyMap)
	for _, parts := range m {
		for i := range parts {
			nk := strings.Join(parts[0:i+1], delim)
			if _, ok := out[nk]; ok {
				continue
			}
			out[nk] = make([]string, i+1)
			copy(out[nk][:], parts[0:i+1])
		}
	}
	return out
}
