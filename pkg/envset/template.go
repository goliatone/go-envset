package envset

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

//DocumentTemplate will create or update a document template
//e.g. envset.tpl that we use to document and to check in our repo
//so we can keep track of the variables and sections.
func DocumentTemplate(name, template string, overwrite, print bool) error {

	filename, err := FileFinder(name)
	if err != nil {
		return err
	}

	ini.PrettyEqual = false
	ini.PrettyFormat = false

	cgf, err := ini.Load(filename)
	if err != nil {
		return err
	}

	var tpl *ini.File
	if overwrite {
		tpl = ini.Empty()
	} else {
		tpl, err = ini.LooseLoad(template)
		if err != nil {
			return err
		}
	}

	for _, sec := range cgf.Sections() {
		tec, err := tpl.GetSection(sec.Name())
		if err != nil {
			tec, err = tpl.NewSection(sec.Name())
			if err != nil {
				return err
			}
		}

		//If our template does not have a comment then we overwrite
		//the comment if the section comment. Note that the source of
		//truth for comments is the template
		if tec.Comment == "" {
			tec.Comment = sec.Comment
		}

		for _, k := range sec.KeyStrings() {
			if tec.HasKey(k) == false {
				//our template has entries MY_VAR={{MY_VAR}}
				_, err := tec.NewKey(k, fmt.Sprintf("{{%s}}", k))
				if err != nil {
					return err
				}
			}

			//Only use .envset comment if template does not have one
			if tec.Key(k).Comment == "" {
				tec.Key(k).Comment = sec.Key(k).Comment
			}
		}

		for _, k := range tec.KeyStrings() {
			if sec.HasKey(k) == false {
				tec.DeleteKey(k)
			}
		}
	}

	if print {
		tpl.WriteTo(os.Stdout)
	} else {

		err = tpl.SaveTo(template) //TODO: This adds two \n at EoF?
		if err != nil {
			return err
		}
	}
	return nil
}
