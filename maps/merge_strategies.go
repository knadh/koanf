package maps

import (
	"fmt"
	"reflect"
)

// MergeAppendSlices recursively merges map src into dest (left to right),
// mutating and expanding dest. Nested maps are merged like Merge. When both
// values at a key are slices, the source slice is appended to the destination
// slice instead of overwriting it. All other conflicting values are overwritten
// by the source value, matching Merge.
//
// Slice element types may differ (for example []string and []any); the result
// is always a fresh []any holding the destination elements followed by the
// source elements. Values taken from src are deep-copied, so mutating src
// after the merge never affects dest. The same guarantee holds for
// MergeByIndex and MergeByKey.
func MergeAppendSlices(src, dest map[string]any) error {
	return mergeWithSliceStrategy(src, dest, appendSlices)
}

// MergeByIndex recursively merges map src into dest (left to right), mutating
// and expanding dest. Nested maps are merged like Merge. When both values at a
// key are slices, elements are merged by index: overlapping map elements are
// deep-merged with the same strategy, overlapping slice elements are merged
// by index recursively, other overlapping elements are overwritten by the
// source, and remaining source elements are appended.
func MergeByIndex(src, dest map[string]any) error {
	return mergeWithSliceStrategy(src, dest, mergeSlicesByIndex)
}

// MergeByKey returns a merge function that behaves like Merge for nested maps
// and non-slice values, but merges slices of maps by matching on the given
// field. For each source map element, if a destination map element shares the
// same field value, the two maps are deep-merged (source into that destination
// element) using the same by-key strategy for nested slices. When several
// destination elements share a field value, the first one is the match.
// Source map elements with no match are appended. Source non-map elements
// overwrite a destination non-map at the same index when present; otherwise
// they are appended.
//
// An empty field name causes the returned function to return an error.
func MergeByKey(field string) func(src, dest map[string]any) error {
	return func(src, dest map[string]any) error {
		if field == "" {
			return fmt.Errorf("merge by key: field name must not be empty")
		}
		return mergeWithSliceStrategy(src, dest, mergeSlicesByKey(field))
	}
}

type sliceStrategy func(srcSlice, destSlice []any) ([]any, error)

func appendSlices(srcSlice, destSlice []any) ([]any, error) {
	out := make([]any, 0, len(destSlice)+len(srcSlice))
	out = append(out, destSlice...)
	for _, el := range srcSlice {
		out = append(out, copyValue(el))
	}
	return out, nil
}

func mergeSlicesByIndex(srcSlice, destSlice []any) ([]any, error) {
	out := make([]any, len(destSlice))
	copy(out, destSlice)
	for i, sv := range srcSlice {
		if i >= len(out) {
			out = append(out, copyValue(sv))
			continue
		}
		sm, sOK := sv.(map[string]any)
		dm, dOK := out[i].(map[string]any)
		if sOK && dOK {
			if err := mergeWithSliceStrategy(sm, dm, mergeSlicesByIndex); err != nil {
				return nil, err
			}
			out[i] = dm
			continue
		}
		if isSlice(sv) && isSlice(out[i]) {
			merged, err := mergeSlicesByIndex(toAnySlice(sv), toAnySlice(out[i]))
			if err != nil {
				return nil, err
			}
			out[i] = merged
			continue
		}
		out[i] = copyValue(sv)
	}
	return out, nil
}

func mergeSlicesByKey(field string) sliceStrategy {
	return func(srcSlice, destSlice []any) ([]any, error) {
		out := make([]any, len(destSlice))
		copy(out, destSlice)

		indexByKey := make(map[any]int)
		for i, el := range out {
			m, ok := el.(map[string]any)
			if !ok {
				continue
			}
			if kv, ok := m[field]; ok {
				// First destination element with a field value wins.
				if _, seen := indexByKey[kv]; !seen {
					indexByKey[kv] = i
				}
			}
		}

		for i, el := range srcSlice {
			sm, ok := el.(map[string]any)
			if !ok {
				if i < len(out) {
					if _, destIsMap := out[i].(map[string]any); !destIsMap {
						out[i] = copyValue(el)
						continue
					}
				}
				out = append(out, copyValue(el))
				continue
			}

			kv, hasKey := sm[field]
			if !hasKey {
				out = append(out, copyValue(el))
				continue
			}

			if di, found := indexByKey[kv]; found {
				dm, ok := out[di].(map[string]any)
				if !ok {
					copied := copyValue(el)
					out[di] = copied
					continue
				}
				if err := mergeWithSliceStrategy(sm, dm, mergeSlicesByKey(field)); err != nil {
					return nil, err
				}
				out[di] = dm
				continue
			}

			out = append(out, copyValue(el))
			indexByKey[kv] = len(out) - 1
		}

		return out, nil
	}
}

func mergeWithSliceStrategy(src, dest map[string]any, mergeSlices sliceStrategy) error {
	for key, val := range src {
		bVal, ok := dest[key]
		if !ok {
			dest[key] = copyValue(val)
			continue
		}

		srcMap, srcIsMap := val.(map[string]any)
		destMap, destIsMap := bVal.(map[string]any)
		if srcIsMap && destIsMap {
			if err := mergeWithSliceStrategy(srcMap, destMap, mergeSlices); err != nil {
				return err
			}
			continue
		}

		if isSlice(val) && isSlice(bVal) {
			merged, err := mergeSlices(toAnySlice(val), toAnySlice(bVal))
			if err != nil {
				return err
			}
			dest[key] = merged
			continue
		}

		dest[key] = copyValue(val)
	}
	return nil
}

func isSlice(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func toAnySlice(v any) []any {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		return nil
	}
	n := rv.Len()
	out := make([]any, n)
	for i := 0; i < n; i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out
}

func copyValue(v any) any {
	if m, ok := v.(map[string]any); ok {
		return Copy(m)
	}
	if isSlice(v) {
		s := toAnySlice(v)
		out := make([]any, len(s))
		for i, el := range s {
			out[i] = copyValue(el)
		}
		return out
	}
	return v
}
