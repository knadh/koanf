module github.com/knadh/koanf-test/providers/posflag

go 1.18

replace (
	github.com/knadh/koanf-test => ../../
	github.com/knadh/koanf-test/maps => ../../maps
)

require (
	github.com/knadh/koanf-test v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/maps v0.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
