package basicflag

// Option is a function that modifies the Pflag
type Option func(*Pflag)

// WithCallBack is a callback function that allows the caller to modify the key and value
func WithCallBack(cb CallBack) Option {
	return func(p *Pflag) {
		p.flagCb = cb
	}
}

// WithEnableMerge It takes a Koanf instance to see if the the flags defined
// have been set from other providers, for instance, a config file.
// If they are not, then the default values of the flags are merged.
// If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func WithEnableMerge(ko KoanfIntf) Option {
	return func(p *Pflag) {
		p.ko = ko
	}
}
