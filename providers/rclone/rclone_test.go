package rclone_test

import (
	encjson "encoding/json"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rclone"
	"github.com/stretchr/testify/assert"
)

func TestRClone(t *testing.T) {
	assert := assert.New(t)

	bRemote, err := os.ReadFile("cloud.txt")
	assert.Nil(err)
	
	remote := string(bRemote)
	remote = strings.TrimSpace(remote)

	f := rclone.Provider(rclone.Config{Remote: remote, File: "parent1.json"})
	assert.Nil(k.Load(f, json.Parser()))

	assert.Equal("json", k.String("type"))
}
