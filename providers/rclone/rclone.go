package rclone

import (
	"context"
	"errors"
	"log"
	"os/user"
	"path/filepath"
	"strings"

	_ "github.com/rclone/rclone/backend/all"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/vfs"
)

type Config struct {
	// Remote storage name from the rclone.conf.
	Remote string

	// RClone configuration file path.
	// Default: /home/${USER}/.config/rclone/rclone.conf
	Cfg string

	// File to read.
	File string
}

type RClone struct {
	Vfs *vfs.VFS
	File string
}

func Provider(cfg Config) *RClone {
	if string(cfg.Remote[len(cfg.Remote)-1]) != ":" {
		cfg.Remote += ":"
	}

	var path string

	usr, _ := user.Current()

	if strings.Compare(cfg.Cfg, "") == 0 {
		path = filepath.Join(usr.HomeDir, ".config/rclone/rclone.conf")
	} else {
		path = cfg.Cfg
	}

	err := config.SetConfigPath(path)
	if err != nil {
		log.Fatal("Error: cannot find RClone config file.\n")
	}

	configfile.Install()

	fs, err := fs.NewFs(context.Background(), cfg.Remote)
	if err != nil {
		log.Fatalf("Error: cannot find remote %s.\n", cfg.Remote)
	}

	vfs := vfs.New(fs, nil)

	return &RClone{Vfs: vfs, File: cfg.File}
}

func (r *RClone) ReadBytes() ([]byte, error) {
	data, err := r.Vfs.ReadFile(r.File)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Read returns the raw bytes for parsing.
func (r *RClone) Read() (map[string]interface{}, error) {
	return nil, errors.New("RClone provider does not support this method")
}

// Watch is not supported.
func (r *RClone) Watch(cb func(event interface{}, err error)) error {
	return errors.New("RClone provider does not support this method")
}
