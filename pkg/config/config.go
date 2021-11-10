package config

import (
	"time"

	"github.com/goliatone/go-envset/pkg/envset"

	"gopkg.in/ini.v1"
)

var config = []byte(`
# Default configuration
filename=.envset
expand=true
isolated=true
export_environment=APP_ENV

[metadata]
dir=.meta
file=data.json
print=true
json=false

[template]
path=.
file=envset.example

[environments]
name=test
name=staging
name=production
name=development
`)

//Config has the rc config options
type Config struct {
	Name         string
	Filename     string `ini:"filename"`
	Environments struct {
		Name []string `ini:"name,omitempty,allowshadow"`
	} `ini:"environments"`
	Created       time.Time `ini:"-"`
	Expand        bool      `ini:"expand"`
	Isolated      bool      `ini:"isolated"`
	ExportEnvName string    `ini:"export_environment"`
	Meta          struct {
		Dir    string `ini:"dir"`
		File   string `ini:"file"`
		Print  bool   `ini:"print"`
		AsJSON bool   `ini:"json"`
	} `ini:"metadata"`
	Template struct {
		Path string `ini:"path"`
		File string `ini:"file"`
	} `ini:"template"`
}

//Load returns configuration object from `.envsetrc` file
func Load(name string) (*Config, error) {
	var err error
	var cfg *ini.File
	var filename string

	filename, _ = envset.FileFinder(name)

	if filename == "" {
		cfg, err = ini.ShadowLoad(config)
	} else {
		cfg, err = ini.ShadowLoad(filename, config)
	}

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

//GetDefaultConfig returns the default
//config string
func GetDefaultConfig() string {
	return string(config)
}
