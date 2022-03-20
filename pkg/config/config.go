package config

import (
	"fmt"
	"path"
	"strconv"
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
	Ignored  map[string][]string
	Required map[string][]string
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

//MergeIgnored will merge ignored values from flags
//with values from envsetrc for a given section
func (c *Config) MergeIgnored(section string, ignored []string) []string {
	if i := c.Ignored[section]; len(i) == 0 {
		return ignored
	}
	out := append(c.Ignored[section], ignored...)
	//TODO: should we make them unique?
	return out
}

//MergeRequired will merge ignored values from flags
//with values from envsetrc for a given section
func (c *Config) MergeRequired(section string, required []string) []string {
	if i := c.Required[section]; len(i) == 0 {
		return required
	}
	out := append(c.Required[section], required...)
	//TODO: should we make them unique?
	return out
}

//Get will return the value of the given key
func (c *Config) Get(key string) string {
	switch key {
	case "filename":
		return c.Filename
	case "environments":
		return printList(c.Environments.Name)
	case "expand":
		return strconv.FormatBool(c.Expand)
	case "isolated":
		return strconv.FormatBool(c.Isolated)
	case "export_environment":
		return c.ExportEnvName
	case "meta.dir":
		return c.Meta.Dir
	case "metadata.dir":
		return c.Meta.Dir
	case "meta.file":
		return c.Meta.File
	case "metadata.file":
		return c.Meta.File
	case "meta.path":
		return path.Join(c.Meta.Dir, c.Meta.File)
	case "metadata.path":
		return path.Join(c.Meta.Dir, c.Meta.File)
	case "template.path":
		return c.Template.Path
	case "template.file":
		return c.Template.File
	case "ignored":
		return printMap(c.Ignored)
	case "required":
		return printMap(c.Required)
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

//ListKeys returns the list of config keys
func (c *Config) ListKeys() []string {
	return []string{
		"filename",
		"environments",
		"expand",
		"isolated",
		"export_environment",
		"meta.dir",
		"meta.file",
		"meta.path",
		"template.path",
		"template.file",
		"ignored",
		"required",
	}
}

//GetDefaultConfig returns the default
//config string
func GetDefaultConfig() string {
	return string(config)
}
