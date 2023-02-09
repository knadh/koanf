module github.com/knadh/koanf/parsers/dotenv

go 1.18

replace (
	github.com/knadh/koanf => ../../
	github.com/knadh/koanf/maps => ../../maps
)

require (
	github.com/joho/godotenv v1.5.1
	github.com/knadh/koanf/maps v0.4.0
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
