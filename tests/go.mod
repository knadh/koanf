module koanf_test

go 1.18

replace (
	github.com/knadh/koanf-test => ../
	github.com/knadh/koanf-test/maps => ../maps
	github.com/knadh/koanf-test/parsers/dotenv => ../parsers/dotenv
	github.com/knadh/koanf-test/parsers/hcl => ../parsers/hcl
	github.com/knadh/koanf-test/parsers/hjson => ../parsers/hjson
	github.com/knadh/koanf-test/parsers/toml => ../parsers/toml
	github.com/knadh/koanf-test/parsers/yaml => ../parsers/yaml
	github.com/knadh/koanf-test/providers/posflag => ../providers/posflag
)

require (
	github.com/knadh/koanf-test v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/parsers/dotenv v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/parsers/hcl v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/parsers/hjson v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/parsers/toml v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/parsers/yaml v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf-test/providers/posflag v0.0.0-00010101000000-000000000000
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hjson/hjson-go/v4 v4.3.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/knadh/koanf-test/maps v0.0.0-00010101000000-000000000000 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20191005200804-aed5e4c7ecf9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
