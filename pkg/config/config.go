package config

import (
	"fmt"
	"goliatone/go-envset/pkg/envset"
	"time"

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
	filename, err := envset.FileFinder(name, 2)
	if err != nil {
		return &Config{}, err
	}

	fmt.Printf("rc: %s\n", filename)

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
