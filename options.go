package koanf

import "github.com/knadh/koanf/maps"

// options contains options to modify the behavior of Koanf.Load.
type options struct {
	merge func(a, b map[string]any) error
}

// newOptions creates a new options instance.
func newOptions(opts []Option) *options {
	o := new(options)
	o.apply(opts)
	return o
}

// Option is a generic type used to modify the behavior of Koanf.Load.
type Option func(*options)

// apply the given options.
func (o *options) apply(opts []Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithMergeFunc is an option to modify the merge behavior of Koanf.Load.
// If unset, the default merge function is used.
//
// The merge function is expected to merge map src into dest (left to right).
func WithMergeFunc(merge func(src, dest map[string]any) error) Option {
	return func(o *options) {
		o.merge = merge
	}
}

// WithMergeAppendSlices makes Load append slices instead of overwriting them
// when merging the incoming source into the existing configuration. Nested maps
// continue to deep-merge. See maps.MergeAppendSlices.
func WithMergeAppendSlices() Option {
	return WithMergeFunc(maps.MergeAppendSlices)
}

// WithMergeByIndex makes Load merge slice elements by index when both sides
// have a slice at the same key. Nested maps continue to deep-merge. See
// maps.MergeByIndex.
func WithMergeByIndex() Option {
	return WithMergeFunc(maps.MergeByIndex)
}

// WithMergeByKey makes Load merge slices of maps by matching on a field when
// both sides have a slice at the same key. Nested maps continue to deep-merge.
// See maps.MergeByKey.
func WithMergeByKey(field string) Option {
	return WithMergeFunc(maps.MergeByKey(field))
}
