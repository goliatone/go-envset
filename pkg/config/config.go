package config

import (
	"time"

	"github.com/goliatone/go-envset/pkg/envset"

	"gopkg.in/ini.v1"
)

var config = []byte(`
filename=.envset
isolated=true

[environments]
name=test
name=staging
name=production
name=development
`)

type Config struct {
	Name         string
	Filename     string `ini:"filename"`
	Environments struct {
		Name []string `ini:"name,omitempty,allowshadow"`
	} `ini:"environments"`
	Created time.Time `ini:"-"`
}

//Load returns configuration object from `.envsetrc` file
func Load(name string) (*Config, error) {
	filename, err := envset.FileFinder(name)
	if err != nil {
		return &Config{}, err
	}

	cfg, err := ini.ShadowLoad(filename, config)
	if err != nil {
		return &Config{}, err
	}

	c := new(Config)
	err = cfg.MapTo(c)

	if err != nil {
		return &Config{}, err
	}

	return c, nil
}
