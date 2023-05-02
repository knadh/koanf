package env

type Option func(*Env)

// WithPrefix sets the environment prefix. Only the env vars with the prefix are captured.
func WithPrefix(prefix string) Option {
	return func(env *Env) {
		env.prefix = prefix
	}
}

// WithDelimiter sets the delimiter to split the environment variable into its parts.
// For example the delimiter "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func WithDelimiter(delim string) Option {
	return func(env *Env) {
		env.delim = delim
	}
}

// WithCallback sets the function that will be called for each environment variable.
// This is useful for cases where you may want to modify the variable or value before it gets passed on.
// If the callback returns an empty string, the variable will be
// ignored.
func WithCallback(cb func(key string, value string) (string, interface{})) Option {
	return func(env *Env) {
		env.cb = cb
	}
}

// WithEnviron sets the environment using a traditional environment slice.
func WithEnviron(environ []string) Option {
	return func(env *Env) {
		env.environ = environ
	}
}

// WithEnvironMap sets the environment using a map.
func WithEnvironMap(environ map[string]string) Option {
	return func(env *Env) {
		env.environ = make([]string, 0, len(environ))
		for k, v := range environ {
			env.environ = append(env.environ, k+"="+v)
		}
	}
}
