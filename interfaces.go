package koanf

// Provider represents a configuration provider. Providers can
// read configuration from a source (file, HTTP etc.)
type Provider interface {
	// Read returns the entire configuration as raw []bytes to be parsed.
	// with a Parser.
	ReadBytes() ([]byte, error)

	// Read returns the prased configuration as a nested map[string]interface{}.
	// It is important to note that the string keys should not be flat delimited
	// keys like `parent.child.key`, but nested like `{parent: {child: {key: 1}}}`.
	Read() (map[string]interface{}, error)

	// Watch watches the source for changes, for instance, changes to a file,
	// and invokes a callback with an `event` interface, which a provider
	// is free to substitute with its own type, including nil.
	Watch(func(event interface{}, err error)) error
}

// Parser represents a configuration format parser.
type Parser interface {
	Parse([]byte) (map[string]interface{}, error)
}
