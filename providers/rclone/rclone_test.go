package rclone_test

import (
	"os"
	"strings"
	"testing"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rclone"
	"github.com/stretchr/testify/assert"
)

func TestRClone(t *testing.T) {
	assert := assert.New(t)
	k := koanf.New(".")

	bRemote, err := os.ReadFile("cloud.txt")
	assert.Nil(err)
	
	remote := string(bRemote)
	remote = strings.TrimSpace(remote)

	f := rclone.Provider(rclone.Config{Remote: remote, File: "mock.json"})
	assert.Nil(k.Load(f, json.Parser()))

	assert.Equal("parent1", k.String("parent1.name"))
	assert.EqualValues(1234, k.Int64("parent1.id"))
	assert.Equal("child1", k.String("parent1.child1.name"))
	assert.Equal(true, k.Bool("parent1.child1.grandchild1.on"))
}
