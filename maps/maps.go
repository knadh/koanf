// Package maps provides reusable functions for manipulating nested
// map[string]interface{} maps are common unmarshal products from
// various serializers such as json, yaml etc.
package maps

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Flatten takes a map[string]interface{} and traverses it and flattens
// nested children into keys delimited by delim.
//
// It's important to note that all nested maps should be
// map[string]interface{} and not map[interface{}]interface{}.
// Use IntfaceKeysToStrings() to convert if necessary.
//
// eg: `{ "parent": { "child": 123 }}` becomes `{ "parent.child": 123 }`
// In addition, it keeps track of and returns a map of the delimited keypaths with
// a slice of key parts, for eg: { "parent.child": ["parent", "child"] }. This
// parts list is used to remember the key path's original structure to
// unflatten later.
func Flatten(m map[string]interface{}, keys []string, delim string) (map[string]interface{}, map[string][]string) {
	var (
		out    = make(map[string]interface{})
		keyMap = make(map[string][]string)
	)
	for key, val := range m {
		// Copy the incoming key paths into a fresh list
		// and append the current key in the iteration.
		kp := make([]string, 0, len(keys)+1)
		kp = append(kp, keys...)
		kp = append(kp, key)

		switch cur := val.(type) {
		case map[string]interface{}:
			// Empty map.
			if len(cur) == 0 {
				newKey := strings.Join(kp, delim)
				out[newKey] = val
				keyMap[newKey] = kp
				continue
			}

			// It's a nested map. Flatten it recursively.
			next, parts := Flatten(cur, kp, delim)

			// Copy the resultant key parts and the value maps.
			for k, p := range parts {
				keyMap[k] = p
			}
			for k, v := range next {
				out[k] = v
			}
		default:
			newKey := strings.Join(kp, delim)
			out[newKey] = val
			keyMap[newKey] = kp
		}
	}
	return out, keyMap
}

// Unflatten takes a flattened key:value map (non-nested with delimited keys)
// and returns a nested map where the keys are split into hierarchies by the given
// delimiter. For instance, `parent.child.key: 1` to `{parent: {child: {key: 1}}}`
//
// It's important to note that all nested maps should be
// map[string]interface{} and not map[interface{}]interface{}.
// Use IntfaceKeysToStrings() to convert if necessary.
func Unflatten(m map[string]interface{}, delim string) map[string]interface{} {
	out := make(map[string]interface{})

	// Iterate through the flat conf map.
	for k, v := range m {
		var (
			keys = strings.Split(k, delim)
			next = out
		)

		// Iterate through key parts, for eg:, parent.child.key
		// will be ["parent", "child", "key"]
		for _, k := range keys[:len(keys)-1] {
			sub, ok := next[k]
			if !ok {
				// If the key does not exist in the map, create it.
				sub = make(map[string]interface{})
				next[k] = sub
			}
			if n, ok := sub.(map[string]interface{}); ok {
				next = n
			}
		}

		// Assign the value.
		next[keys[len(keys)-1]] = v
	}
	return out
}

// Merge recursively merges map a into b (left to right), mutating
// and expanding map b. Note that there's no copying involved, so
// map b will retain references to map a.
//
// It's important to note that all nested maps should be
// map[string]interface{} and not map[interface{}]interface{}.
// Use IntfaceKeysToStrings() to convert if necessary.
func Merge(a, b map[string]interface{}) {
	for key, val := range a {
		// Does the key exist in the target map?
		// If no, add it and move on.
		bVal, ok := b[key]
		if !ok {
			b[key] = val
			continue
		}

		// If the incoming val is not a map, do a direct merge.
		if _, ok := val.(map[string]interface{}); !ok {
			b[key] = val
			continue
		}

		// The source key and target keys are both maps. Merge them.
		switch v := bVal.(type) {
		case map[string]interface{}:
			Merge(val.(map[string]interface{}), v)
		default:
			b[key] = val
		}
	}
}

// Search recursively searches a map for a given path. The path is
// the key map slice, for eg:, parent.child.key -> [parent child key].
//
// It's important to note that all nested maps should be
// map[string]interface{} and not map[interface{}]interface{}.
// Use IntfaceKeysToStrings() to convert if necessary.
func Search(mp map[string]interface{}, path []string) interface{} {
	next, ok := mp[path[0]]
	if ok {
		if len(path) == 1 {
			return next
		}
		switch next.(type) {
		case map[string]interface{}:
			return Search(next.(map[string]interface{}), path[1:])
		default:
			return nil
		} //
		// It's important to note that all nested maps should be
		// map[string]interface{} and not map[interface{}]interface{}.
		// Use IntfaceKeysToStrings() to convert if necessary.
	}
	return nil
}

// Copy returns a copy of a conf map by doing a JSON marshal+unmarshal
// pass. Inefficient, but creates a true deep copy. There is a side
// effect, that is, all numeric types change to float64.
//
// It's important to note that all nested maps should be
// map[string]interface{} and not map[interface{}]interface{}.
// Use IntfaceKeysToStrings() to convert if necessary.
func Copy(mp map[string]interface{}) map[string]interface{} {
	var out map[string]interface{}
	b, _ := json.Marshal(mp)
	json.Unmarshal(b, &out)
	return out
}

// IntfaceKeysToStrings recursively converts map[interface{}]interface{} to
// map[string]interface{}. Some parses such as YAML unmarshal return this.
func IntfaceKeysToStrings(mp map[string]interface{}) {
	for key, val := range mp {
		switch cur := val.(type) {
		case map[interface{}]interface{}:
			x := make(map[string]interface{})
			for k, v := range cur {
				x[fmt.Sprintf("%v", k)] = v
			}
			mp[key] = x
			IntfaceKeysToStrings(x)
		case []interface{}:
			for i, v := range cur {
				switch sub := v.(type) {
				case map[interface{}]interface{}:
					x := make(map[string]interface{})
					for k, v := range sub {
						x[fmt.Sprintf("%v", k)] = v
					}
					cur[i] = x
					IntfaceKeysToStrings(x)
				case map[string]interface{}:
					IntfaceKeysToStrings(sub)
				}
			}
		case map[string]interface{}:
			IntfaceKeysToStrings(cur)
		}
	}
}

// StringSliceToLookupMap takes a slice of strings and returns a lookup map
// with the slice values as keys with true values.
func StringSliceToLookupMap(s []string) map[string]bool {
	mp := make(map[string]bool, len(s))
	for _, v := range s {
		mp[v] = true
	}
	return mp
}

// Int64SliceToLookupMap takes a slice of int64s and returns a lookup map
// with the slice values as keys with true values.
func Int64SliceToLookupMap(s []int64) map[int64]bool {
	mp := make(map[int64]bool, len(s))
	for _, v := range s {
		mp[v] = true
	}
	return mp
}
