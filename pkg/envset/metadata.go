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
	"time"

	"gopkg.in/ini.v1"
)

const (
	//HashSHA256 is the default hash algorithm
	HashSHA256 = "sha256"
	//HashHMAC is the algorithm used with a secret
	HashHMAC = "hmac"
	//HashMD5 used md5 hash
	HashMD5 = "md5"
)

//EnvFile struct
type EnvFile struct {
	//TODO: make relative to executable
	Path      string        `json:"-"`
	File      *ini.File     `json:"-"`
	Filename  string        `json:"envfile,omitempty"`
	Project   string        `json:"project,omitempty"` //TODO: should we have https, git and id? if someone checks using https and other ssh this will change!!
	Algorithm string        `json:"algorithm"`
	Date      time.Time     `json:"date"`
	Sections  []*EnvSection `json:"sections"` //TODO: make custom marshaller to ignore DEFAULT section
	secret    string
}

//EnvSection is a top level section
type EnvSection struct {
	Name      string    `json:"name"`
	Comment   string    `json:"comment,omitempty"`
	Keys      []*EnvKey `json:"values"`
	secret    string
	algorithm string
	maxLength int
}

//AddKey adds a new key to the section
func (e *EnvSection) AddKey(key, value string) (*EnvKey, error) {

	hash, err := e.makeHash(key, value)

	if err != nil {
		return &EnvKey{}, err
	}

	envKey := &EnvKey{
		Name:  key,
		Value: value,
		Hash:  hash,
	}

	e.Keys = append(e.Keys, envKey)

	return envKey, nil
}

func (e *EnvSection) makeHash(key, value string) (string, error) {
	var err error
	var hash string
	fmt.Printf("make hash: %s\n", e.algorithm)
	switch e.algorithm {
	case HashHMAC:
		hash, err = hmacSha256HashValue(value, e.secret)
	case HashSHA256:
		hash, err = sha256Hashvalue(value)
	case HashMD5:
		hash, err = md5HashValue(value)
	default:
		hash, err = sha256Hashvalue(value)
	}

	if len(hash) > e.maxLength {
		hash = hash[:e.maxLength]
	}

	return hash, err
}

//IsEmpty will return true if we have no keys in our section
func (e *EnvSection) IsEmpty() bool {
	return len(e.Keys) == 0
}

//ToJSON returns a JSON representation of a section
func (e *EnvSection) ToJSON() (string, error) {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return "", fmt.Errorf("section json marshall: %w", err)
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
		Name:      name,
		Keys:      make([]*EnvKey, 0),
		algorithm: e.Algorithm,
		secret:    e.secret,
		maxLength: 50,
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
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return "", fmt.Errorf("file json marshall: %w", err)
	}
	return string(b), nil
}

//FromJSON load from json file
func (e *EnvFile) FromJSON(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %s: %w", path, err)
	}

	err = json.Unmarshal([]byte(file), &e)
	if err != nil {
		return fmt.Errorf("unmarshal file %s: %w", path, err)
	}

	return nil
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
func CreateMetadataFile(o MetadataOptions) (EnvFile, error) {

	filename, err := FileFinder(o.Name)
	if err != nil {
		return EnvFile{}, fmt.Errorf("file finder: %w", err)
	}

	ini.PrettyEqual = false
	ini.PrettyFormat = false
	fmt.Printf("create meta: %s\n", o.Algorithm)
	algorithm := o.Algorithm
	if o.Secret != "" {
		algorithm = HashHMAC
	}

	envFile := EnvFile{
		Algorithm: algorithm,
		Path:      filename,
		Filename:  o.Name,
		Project:   o.Project,
		Sections:  make([]*EnvSection, 0),
		Date:      time.Now().UTC(),
		secret:    o.Secret,
	}

	cfg, err := ini.Load(filename)
	if err != nil {
		return EnvFile{}, fmt.Errorf("ini load %s: %w", filename, err)
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
			envKey, err := envSect.AddKey(k, v)

			if o.Values == false {
				envKey.Value = ""
			}

			if err != nil {
				return EnvFile{}, fmt.Errorf("add key %s=%s: %w", k, v, err)
			}

			if sec.Key(k).Comment != "" {
				envKey.Comment = sec.Key(k).Comment
			}
		}
	}

	return envFile, nil
}

//LoadMetadataFile will load a metadata file from the provided path
func LoadMetadataFile(path string) (*EnvFile, error) {
	envFile := &EnvFile{}
	err := envFile.FromJSON(path)
	if err != nil {
		return nil, fmt.Errorf("load metadata file: %w", err)
	}
	return envFile, nil
}

//CompareMetadataFiles will compare two EnvFile instances
func CompareMetadataFiles(a, b *EnvFile) (bool, error) {

	if len(a.Sections) != len(b.Sections) {
		return true, nil
	}

	for _, sa := range a.Sections {
		for _, sb := range b.Sections {
			if sa.Name != sb.Name {
				continue
			}
			diff := CompareSections(*sa, *sb, []string{})
			if !diff.IsEmpty() {
				return true, nil
			}
			break
		}
	}

	for _, sb := range b.Sections {
		for _, sa := range a.Sections {
			if sa.Name != sb.Name {
				continue
			}
			diff := CompareSections(*sa, *sb, []string{})
			if !diff.IsEmpty() {
				return true, nil
			}
			break
		}
	}

	return false, nil
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

//CompareSections will compare two sections and return diff
func CompareSections(s1, s2 EnvSection, ignored []string) EnvSection {
	ignore := make(map[string]bool)
	for _, v := range ignored {
		ignore[v] = true
	}

	diff := EnvSection{}
	seen := make(map[string]int)

	for i, k1 := range s1.Keys {
		if ok := ignore[k1.Name]; ok {
			//TODO: diff.Ignored = append(diff.Ignored, k1)
			continue
		}

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
		if ok := ignore[k2.Name]; ok {
			//TODO: diff.Ignored = append(diff.Ignored, k2)
			continue
		}

		if _, ok := seen[k2.Name]; ok == false {
			k2.Comment = "missing in source"
			diff.Keys = append(diff.Keys, k2)
		}
	}

	return diff
}
