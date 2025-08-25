module github.com/knadh/koanf/koanf_tests

go 1.23.0

replace (
	github.com/knadh/koanf/maps => ../maps
	github.com/knadh/koanf/parsers/dotenv => ../parsers/dotenv
	github.com/knadh/koanf/parsers/hcl => ../parsers/hcl
	github.com/knadh/koanf/parsers/hjson => ../parsers/hjson
	github.com/knadh/koanf/parsers/json => ../parsers/json
	github.com/knadh/koanf/parsers/toml => ../parsers/toml
	github.com/knadh/koanf/parsers/yaml => ../parsers/yaml
	github.com/knadh/koanf/providers/basicflag => ../providers/basicflag
	github.com/knadh/koanf/providers/confmap => ../providers/confmap
	github.com/knadh/koanf/providers/env/v2 => ../providers/env
	github.com/knadh/koanf/providers/file => ../providers/file
	github.com/knadh/koanf/providers/fs => ../providers/fs
	github.com/knadh/koanf/providers/posflag => ../providers/posflag
	github.com/knadh/koanf/providers/rawbytes => ../providers/rawbytes
	github.com/knadh/koanf/providers/structs => ../providers/structs
	github.com/knadh/koanf/v2 => ../
)

require (
	github.com/knadh/koanf/maps v0.1.2
	github.com/knadh/koanf/parsers/dotenv v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/parsers/hcl v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/parsers/hjson v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/parsers/json v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/parsers/toml v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/parsers/yaml v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/basicflag v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/confmap v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/env/v2 v2.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/file v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/fs v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/posflag v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/rawbytes v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/providers/structs v0.0.0-00010101000000-000000000000
	github.com/knadh/koanf/v2 v2.0.0-00010101000000-000000000000
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hjson/hjson-go/v4 v4.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/sys v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
