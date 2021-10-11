package envset

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/ini.v1"
)

//EnvFile struct
type EnvFile struct {
	//TODO: make relative to executable
	Path     string    `json:"-"`
	File     *ini.File `json:"-"`
	Filename string    `json:"envfile,omitempty"`
	//TODO: should we have https, git and id? if someone checks using
	//https and other ssh this will change!!
	Project string `json:"project,omitempty"`
	Alg     string `json:"algorithm"`
	//TODO: make custom marshaller to ignore DEFAULT section
	Sections []*EnvSection `json:"sections"`
}

//EnvSection is a top level section
type EnvSection struct {
	Name    string    `json:"name"`
	Comment string    `json:"comment,omitempty"`
	Keys    []*EnvKey `json:"values"`
}

//AddKey adds a new key to the section
func (e *EnvSection) AddKey(key, value, secret string) (*EnvKey, error) {
	var err error
	var hash string

	//TODO: Consider removing no secret option
	if secret == "" {
		hash, err = sha256Hashvalue(value)
	} else {
		hash, err = hmacSha256HashValue(value, secret)
	}

	if err != nil {
		return &EnvKey{}, err
	}

	hash = hash[:50]

	envKey := &EnvKey{
		Name:  key,
		Value: value,
		Hash:  hash,
	}

	e.Keys = append(e.Keys, envKey)

	return envKey, nil
}

//IsEmpty will return true if we have no keys in our section
func (e *EnvSection) IsEmpty() bool {
	return len(e.Keys) == 0
}

//ToJSON returns a JSON representation of a section
func (e *EnvSection) ToJSON() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
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
		return fmt.Errorf("ini load: %w", err)
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

//GetSection will return a EnvSection by name or an error if is
//not found
func (e *EnvFile) GetSection(name string) (*EnvSection, error) {

	for _, section := range e.Sections {
		if section.Name == name {
			return section, nil
		}
	}
	return &EnvSection{}, errors.New("Section not found")
}

//ToJSON will print the JSON representation for a envfile
func (e EnvFile) ToJSON() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("json marshall: %w", err)
	}
	return string(b), nil
}

//FromJSON load from json file
func (e *EnvFile) FromJSON(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %s: %w", path, err)
	}

	return json.Unmarshal([]byte(file), &e)
}

//FromStdin read from stdin
func (e *EnvFile) FromStdin() error {
	return json.NewDecoder(os.Stdin).Decode(&e)
}

//MetadataOptions are the command options
type MetadataOptions struct {
	Name          string
	Filepath      string
	Algorithm     string
	Project       string
	GlobalSection string
	Overwrite     bool
	Globals       bool
	Print         bool
	Values        bool
	Secret        string
}

//CreateMetadataFile will create or update metadata file
func CreateMetadataFile(o MetadataOptions) error {

	filename, err := FileFinder(o.Name)
	if err != nil {
		return fmt.Errorf("file finder: %w", err)
	}

	ini.PrettyEqual = false
	ini.PrettyFormat = false

	envFile := EnvFile{
		Alg:      o.Algorithm,
		Path:     filename,
		Filename: o.Name,
		Project:  o.Project,
		Sections: make([]*EnvSection, 0),
	}

	// err = envFile.Load(filename)
	cfg, err := ini.Load(filename)
	if err != nil {
		return fmt.Errorf("ini load %s: %w", filename, err)
	}

	for _, sec := range cfg.Sections() {

		secName := sec.Name()

		//Check for defaults sections
		if secName == ini.DEFAULT_SECTION {
			if len(sec.KeyStrings()) == 0 || o.Globals == false {
				continue
			}
		}

		//Add section e.g. [development]
		envSect := envFile.AddSection(secName)

		if sec.Comment != "" {
			envSect.Comment = sec.Comment
		}

		//Go over section and add new EnvKeys
		for _, k := range sec.KeyStrings() {
			v := sec.Key(k).String()
			envKey, err := envSect.AddKey(k, v, o.Secret)

			if o.Values == false {
				envKey.Value = ""
			}

			if err != nil {
				return fmt.Errorf("add key %s=%s: %w", k, v, err)
			}

			if sec.Key(k).Comment != "" {
				envKey.Comment = sec.Key(k).Comment
			}
		}
	}

	str, err := envFile.ToJSON()
	if err != nil {
		return fmt.Errorf("env file to json: %w", err)
	}

	if o.Print {
		fmt.Print(str)
	} else {
		if _, err := os.Stat(o.Filepath); os.IsNotExist(err) {
			err := ioutil.WriteFile(o.Filepath, []byte(str), 0777)
			if err != nil {
				return fmt.Errorf("write file %s: %w", o.Filepath, err)
			}
		} else if o.Overwrite == true {
			err := ioutil.WriteFile(o.Filepath, []byte(str), 0777)
			if err != nil {
				return fmt.Errorf("overwrite file %s: %w", o.Filepath, err)
			}
		}
	}

	return nil
}

func md5HashValue(value string) (string, error) {
	hash := md5.New()
	hash.Write([]byte(value))
	sha := fmt.Sprintf("%x", hash.Sum(nil))
	return sha, nil
}

func sha256Hashvalue(value string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(value))
	sha := fmt.Sprintf("%x", hash.Sum(nil))
	return sha, nil
}

func hmacSha256HashValue(value, secret string) (string, error) {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(value))
	sha := hex.EncodeToString(hash.Sum(nil))
	return sha, nil
}

//https://golang.org/pkg/crypto/hmac/
// func ValidMAC(message, messageMAC, key []byte) bool {
// 	mac := hmac.New(sha256.New, key)
// 	mac.Write(message)
// 	expectedMAC := mac.Sum(nil)
// 	return hmac.Equal(messageMAC, expectedMAC)
// }

//CompareSections will compare two sections and return diff
func CompareSections(s1, s2 EnvSection) EnvSection {
	diff := EnvSection{}
	seen := make(map[string]int)

	for i, k1 := range s1.Keys {
		seen[k1.Name] = -1
		for _, k2 := range s2.Keys {
			if k1.Name == k2.Name {
				seen[k1.Name] = i + 1
				if k1.Hash != k2.Hash {
					k1.Comment = "different hash value"
					diff.Keys = append(diff.Keys, k1)
					break
				}
			}
		}

		if seen[k1.Name] == -1 {
			k1.Comment = "extra in source"
			diff.Keys = append(diff.Keys, k1)
		}
	}

	for _, k2 := range s2.Keys {
		if _, ok := seen[k2.Name]; ok == false {
			k2.Comment = "missing in source"
			diff.Keys = append(diff.Keys, k2)
		}
	}

	return diff
}
