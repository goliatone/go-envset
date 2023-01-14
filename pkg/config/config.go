package config

import (
	"fmt"
	"path"
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
restart=true
max_restarts=3
restart_forever=false
restart_exclude=test

[metadata]
dir=.meta
file=data.json
print=true
json=false

[template]
dir=.
file=envset.example

[environments]
name=test
name=staging
name=production
name=development

[comments]
key=COMMENTS
key=DOCUMENTATION
`)

// Config has the rc config options
type Config struct {
	Name                string
	Filename            string               `ini:"filename"`
	Environments        *Environments        `ini:"environments"`
	CommentSectionNames *CommentSectionNames `ini:"comments"`
	Created             time.Time            `ini:"-"`
	Expand              bool                 `ini:"expand"`
	Isolated            bool                 `ini:"isolated"`
	ExportEnvName       string               `ini:"export_environment"`
	ExportEnvNameOld    string               `ini:"exportEnvironment"`
	Meta                *Meta                `ini:"metadata"`
	Template            *Template            `ini:"template"`
	Ignored             map[string][]string
	Required            map[string][]string
	Restart             bool     `ini:"restart"`
	ExcludeFromRestart  []string `ini:"restart_exclude"`
	RestartForever      bool     `ini:"restart_forever"`
	MaxRestarts         int      `ini:"max_restarts"`
}

// Environments holds the environment names
type Environments struct {
	Names []string `ini:"name,omitempty,allowshadow"`
}

// CommentSectionNames are the sections that will be ignored
type CommentSectionNames struct {
	Keys []string `ini:"key,omitempty,allowshadow"`
}

// Meta are options for the metadata command
type Meta struct {
	Dir    string `ini:"dir"`
	File   string `ini:"file"`
	Print  bool   `ini:"print"`
	AsJSON bool   `ini:"json"`
}

// Template are options to generate the template output
type Template struct {
	Dir  string `ini:"dir"`
	File string `ini:"file"`
}

// Load returns configuration object from `.envsetrc` file
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

	c := newConfig()
	err = cfg.MapTo(c)
	if err != nil {
		return &Config{}, err
	}

	if c.ExportEnvNameOld != "" {
		c.ExportEnvName = c.ExportEnvNameOld
	}

	if sec, err := cfg.GetSection("ignored"); err == nil {
		c.Ignored = make(map[string][]string)
		for _, k := range sec.KeyStrings() {
			v := sec.Key(k).ValueWithShadows()
			c.Ignored[k] = v
		}
	}

	if sec, err := cfg.GetSection("required"); err == nil {
		c.Required = make(map[string][]string)
		for _, k := range sec.KeyStrings() {
			v := sec.Key(k).ValueWithShadows()
			c.Required[k] = v
		}
	}

	return c, nil
}

// MergeIgnored will merge ignored values from flags
// with values from envsetrc for a given section
func (c *Config) MergeIgnored(section string, ignored []string) []string {
	if i := c.Ignored[section]; len(i) == 0 {
		return ignored
	}
	out := append(c.Ignored[section], ignored...)
	//TODO: should we make them unique?
	return out
}

// MergeRequired will merge ignored values from flags
// with values from envsetrc for a given section
func (c *Config) MergeRequired(section string, required []string) []string {
	if i := c.Required[section]; len(i) == 0 {
		return required
	}
	out := append(c.Required[section], required...)
	//TODO: should we make them unique?
	return out
}

// Get will return the value of the given key
func (c *Config) Get(key string) string {
	switch key {
	case "filename":
		return c.Filename
	case "meta.dir":
		return c.Meta.Dir
	case "metadata.dir":
		return c.Meta.Dir
	case "meta.file":
		return c.Meta.File
	case "metadata.file":
		return c.Meta.File
	case "meta.filepath":
		return path.Join(c.Meta.Dir, c.Meta.File)
	case "metadata.filepath":
		return path.Join(c.Meta.Dir, c.Meta.File)
	case "template.dir":
		return c.Template.Dir
	case "template.file":
		return c.Template.File
	case "template.filepath":
		return path.Join(c.Template.Dir, c.Template.File)
	default:
		return ""
	}
}

func printMap(s map[string][]string) string {
	o := ""
	for k, m := range s {
		l := printList(m)
		o += fmt.Sprintf("%s - \n%s", k, l)
	}
	return o
}

func printList(a []string) string {
	o := ""
	for _, s := range a {
		o += fmt.Sprintf("%s\n", s)
	}
	return o
}

// ListKeys returns the list of config keys
func (c *Config) ListKeys() []string {
	return []string{
		"filename",
		"meta.dir",
		"meta.file",
		"meta.filepath",
		"template.dir",
		"template.file",
		"template.filepath",
		"ignored",
		"required",
	}
}

// RestartForEnv will return the restart value based on the env
func (c *Config) RestartForEnv(env string) bool {
	for _, exc := range c.ExcludeFromRestart {
		if exc == env {
			return false
		}
	}
	return c.Restart
}

// GetDefaultConfig returns the default
// config string
func GetDefaultConfig() string {
	return string(config)
}

func newConfig() *Config {
	c := new(Config)
	c.Filename = ".envset"
	c.Expand = true
	c.Isolated = true
	c.ExportEnvName = "APP_ENV"
	c.Restart = true
	c.MaxRestarts = 3
	c.RestartForever = false
	c.Meta = &Meta{
		Dir:    ".meta",
		File:   "data.json",
		Print:  true,
		AsJSON: false,
	}
	c.Template = &Template{
		Dir:  ".",
		File: "envset.example",
	}
	c.Environments = &Environments{
		Names: []string{
			"development",
			"staging",
			"production",
			"test",
		},
	}
	return c
}
