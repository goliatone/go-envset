package envset

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// DocumentTemplate will create or update a document template
// e.g. envset.tpl that we use to document and to check in our repo
// so we can keep track of the variables and sections.
func DocumentTemplate(name, template string, overwrite, printOutput bool) error {
	filename, err := FileFinder(name)
	if err != nil {
		return fmt.Errorf("file finder %s: %w", name, err)
	}

	ini.PrettyEqual = false
	ini.PrettyFormat = false

	cgf, err := ini.Load(filename)
	if err != nil {
		return fmt.Errorf("ini load %s: %w", filename, err)
	}

	tpl, err := loadTemplateFile(template, overwrite)
	if err != nil {
		return err
	}

	for _, sec := range cgf.Sections() {
		if err := syncTemplateSection(tpl, sec); err != nil {
			return err
		}
	}

	return writeTemplateFile(tpl, template, printOutput)
}

func loadTemplateFile(template string, overwrite bool) (*ini.File, error) {
	if overwrite {
		return ini.Empty(), nil
	}

	tpl, err := ini.LooseLoad(template)
	if err != nil {
		return nil, fmt.Errorf("ini loose load %s: %w", template, err)
	}
	return tpl, nil
}

func syncTemplateSection(tpl *ini.File, sec *ini.Section) error {
	tec, err := getOrCreateTemplateSection(tpl, sec.Name())
	if err != nil {
		return err
	}

	if tec.Comment == "" {
		tec.Comment = sec.Comment
	}

	for _, key := range sec.KeyStrings() {
		if err := syncTemplateKey(tec, sec, key); err != nil {
			return err
		}
	}

	removeMissingTemplateKeys(tec, sec)
	return nil
}

func getOrCreateTemplateSection(tpl *ini.File, name string) (*ini.Section, error) {
	sec, err := tpl.GetSection(name)
	if err == nil {
		return sec, nil
	}

	sec, err = tpl.NewSection(name)
	if err != nil {
		return nil, fmt.Errorf("tpl new section %s: %w", name, err)
	}
	return sec, nil
}

func syncTemplateKey(tec, sec *ini.Section, key string) error {
	if !tec.HasKey(key) {
		if _, err := tec.NewKey(key, fmt.Sprintf("{{%s}}", key)); err != nil {
			return fmt.Errorf("new key %s: %w", key, err)
		}
	}

	if tec.Key(key).Comment == "" {
		tec.Key(key).Comment = sec.Key(key).Comment
	}
	return nil
}

func removeMissingTemplateKeys(tec, sec *ini.Section) {
	for _, key := range tec.KeyStrings() {
		if !sec.HasKey(key) {
			tec.DeleteKey(key)
		}
	}
}

func writeTemplateFile(tpl *ini.File, template string, printOutput bool) error {
	if printOutput {
		if _, err := tpl.WriteTo(os.Stdout); err != nil {
			return fmt.Errorf("tpl write stdout: %w", err)
		}
		return nil
	}

	if err := tpl.SaveTo(template); err != nil {
		return fmt.Errorf("tpl save to %s: %w", template, err)
	}
	return nil
}
