package envset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"crypto/md5"
	"gopkg.in/ini.v1"
)
//EnvFile struct
type EnvFile struct {
	Path      string `json:"path"`
	File      *ini.File  `json:"-"`
	Sections  []*EnvSection `json:"sections"`
}

//EnvSection is a top level section
type EnvSection struct {
	Name    string   `json:"name"`
	// Comment string   `json:"comment"`
	Keys    []*EnvKey `json:"values"`
}

func (e *EnvSection) AddKey(key, value string) error {
	hash, err := md5HashValue(value)
	if err != nil {
		return err
	}
	
	e.Keys = append(e.Keys, &EnvKey{
		Name: key, 
		Value: value, 
		Hash: hash,
	})

	return nil
}

//EnvKey is a single entry in our file
type EnvKey struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Hash    string `json:"hash"`
}

//Load will load the ini file
func (e EnvFile) Load(path string) error {
	file, err := ini.Load(path)
	if err != nil {
		return err
	}

	e.Path = path
	e.File = file

	return nil
}

//AddSection will add a section to a EnvFile
func (e *EnvFile) AddSection(name string) *EnvSection {
	es := &EnvSection{
		Name: name,
		Keys: make([]*EnvKey, 0),
	}
	e.Sections = append(e.Sections, es)
	return es
}

//ToJSON will print the JSON representation for a envfile
func (e EnvFile) ToJSON() (string, error) {
	b, err := json.Marshal(e)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

//CreateMetadataFile will create or update metadata file
func CreateMetadataFile(name, filepath string, overwrite, print, hash bool) error {

	filename, err := FileFinder(name)
	if err != nil {
		return err
	}

	ini.PrettyEqual = false
	ini.PrettyFormat = false

	envFile := EnvFile{
		Path: filename,
		Sections: make([]*EnvSection, 0),
	}

	// err = envFile.Load(filename)
	cfg, err := ini.Load(filename)
	if err != nil {
		return err
	}

	for _, sec := range cfg.Sections() {
		//Add section [development]
		envSect := envFile.AddSection(sec.Name())

		//Go over section and add new EnvKeys
		for _, k := range sec.KeyStrings() {
			v := sec.Key(k).String()
			envSect.AddKey(k, v)
		}
	}

	str, err := envFile.ToJSON()
	if err != nil {
		return err
	}

	if print {
		fmt.Print(str)
	} else {
		err := ioutil.WriteFile(filepath, []byte(str), 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func md5HashValue(value string) (string, error) {
	hash := md5.New()
	hash.Write([]byte(value))
	str := fmt.Sprintf("%x", hash.Sum(nil))
	return str, nil
}