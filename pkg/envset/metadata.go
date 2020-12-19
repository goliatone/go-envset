package envset

import (
	"encoding/json"
	"fmt"
	"crypto/md5"
	"io/ioutil"
	"os"

	"gopkg.in/ini.v1"
)
//EnvFile struct
type EnvFile struct {
	//TODO: make relative to executable
	Path      string     	`json:"-"`
	File      *ini.File  	`json:"-"`
	Filename  string     	`json:"envfile,omitempty"`
	//TODO: should we have https, git and id? if someone checks using 
	//https and other ssh this will change!!
	Project   string     	`json:"project,omitempty"`
	Alg 	  string 		`json:"algorithm"`
	//TODO: make custom marshaller to ignore DEFAULT section
	Sections  []*EnvSection `json:"sections"`
}

//EnvSection is a top level section
type EnvSection struct {
	Name    string    `json:"name"`
	Comment string    `json:"comment,omitempty"`
	Keys    []*EnvKey `json:"values"`
}

func (e *EnvSection) AddKey(key, value string) (*EnvKey, error) {
	hash, err := md5HashValue(value)
	if err != nil {
		return &EnvKey{}, err
	}
	
	envKey := &EnvKey{
		Name: key, 
		Value: value, 
		Hash: hash,
	}

	e.Keys = append(e.Keys, envKey)

	return envKey, nil
}

//EnvKey is a single entry in our file
type EnvKey struct {
	Name    string `json:"key"`
	Value   string `json:"value,omitempty"`
	Hash    string `json:"hash"`
	Comment string `json:"comment,omitempty"`
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

type MetadataOptions struct {
	Name 	  string 
	Filepath  string
	Algorithm string
	Project   string
	Overwrite bool 
	Print 	  bool 
	Values 	  bool
}

//CreateMetadataFile will create or update metadata file
//TODO: we need to add project name, should match repo
// func CreateMetadataFile(name, filepath string, overwrite, print, hash bool) error {
func CreateMetadataFile(o MetadataOptions) error {

	filename, err := FileFinder(o.Name)
	if err != nil {
		return err
	}

	ini.PrettyEqual = false
	ini.PrettyFormat = false

	envFile := EnvFile{
		Alg: o.Algorithm,
		Path: filename,
		Filename: o.Name,
		Project: o.Project,
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

		if sec.Comment != "" {
			envSect.Comment = sec.Comment
		}

		//Go over section and add new EnvKeys
		for _, k := range sec.KeyStrings() {
			v := sec.Key(k).String()
			envKey, err := envSect.AddKey(k, v)
			
			if o.Values == false {
				envKey.Value = ""
			}

			if err != nil {
				return err
			}
			if sec.Key(k).Comment != "" {
				envKey.Comment = sec.Key(k).Comment
			}
		}
	}

	str, err := envFile.ToJSON()
	if err != nil {
		return err
	}

	if o.Print {
		fmt.Print(str)
	} else {
		//TODO: check to see if file exists and we have o.Overwrite false
		if _, err := os.Stat(o.Filepath); os.IsNotExist(err) {
			err := ioutil.WriteFile(o.Filepath, []byte(str), 0777)
			if err != nil {
				return err
			}
		} else if o.Overwrite == true {
			err := ioutil.WriteFile(o.Filepath, []byte(str), 0777)
			if err != nil {
				return err
			}
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