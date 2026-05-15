package metadata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goliatone/go-envset/cmd/envset/internal/cliopts"
	"github.com/goliatone/go-envset/pkg/config"
	"github.com/goliatone/go-envset/pkg/envset"
	"github.com/gosuri/uitable"
	colors "github.com/logrusorgru/aurora/v3"
	"github.com/tcnksm/go-gitconfig"
	"github.com/urfave/cli/v2"
)

// GetCommand returns a new cli.Command for the
// metadata command.
func GetCommand(cnf *config.Config) *cli.Command {

	return &cli.Command{
		Name:        "metadata",
		Usage:       "generate a metadata file from environment file",
		Description: "creates a metadata file with all the given environments",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "print", Usage: "only print the contents to stdout, don't write file"},
			&cli.StringFlag{Name: "filename", Usage: "metadata file `name`", Value: cnf.Meta.File},
			&cli.StringFlag{Name: "filepath", Usage: "metadata file `path`", Value: cnf.Meta.Dir},
			&cli.StringFlag{Name: "env-file", Value: cnf.Filename, Usage: "load environment from `FILE`"},
			&cli.BoolFlag{Name: "overwrite", Usage: "set to false to prevent overwrite metadata file", Value: true},
			&cli.BoolFlag{Name: "values", Usage: "add flag to show values in the output"},
			&cli.BoolFlag{Name: "globals", Usage: "include global section", Value: false},
			&cli.StringFlag{Name: "secret", Usage: "`password` used to encode hash values. Define env ENVSET_HASH_SECRET", EnvVars: []string{"ENVSET_HASH_SECRET"}},
			&cli.StringFlag{
				Name:    "hash-algo",
				Usage:   "hash algorithm used to hash values. Define with ENVSET_HASH_ALGORITHM",
				EnvVars: []string{"ENVSET_HASH_ALGORITHM"},
				Value:   envset.HashSHA256,
			},
		},
		Action: runMetadataCommand,
		Subcommands: []*cli.Command{
			{
				Name:  "compare",
				Usage: "compare two metadata files",
				UsageText: `envset metadata compare --section=[section] [target]
   envset metadata compare --section=[section] [source] [target]

EXAMPLE:
   envset metadata compare --section=development .meta/data.json .meta/prod.data.json`,
				Description: `compares the provided [section] of two metadata files
   [source] by default is .meta/data.json`,
				Category: "METADATA",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "section",
						Aliases:  []string{"s"},
						Usage:    "env file section",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "print",
						Usage: "print the comparison results to stdout",
						Value: cnf.Meta.Print,
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "print the comparison results to stdout in JSON format",
						Value: cnf.Meta.AsJSON,
					},
					&cli.StringSliceFlag{
						Name:    "ignore",
						Aliases: []string{"I"},
						Usage:   "list of key names that are ignored",
					},
				},
				Action: func(c *cli.Context) error {
					return runMetadataCompare(cnf, c)
				},
			},
		},
	}
}

func runMetadataCommand(c *cli.Context) error {
	options, dir, shouldClean, err := metadataOptions(c)
	if err != nil {
		return err
	}

	newEnv, err := envset.CreateMetadataFile(options)
	if err != nil {
		return err
	}

	envExists := exists(options.Filepath)
	if envExists {
		changed, err := metadataChanged(options.Filepath, &newEnv)
		if err != nil || !changed && !options.Print {
			return err
		}
	}

	contents, err := newEnv.ToJSON()
	if err != nil {
		return fmt.Errorf("env file to json: %w", err)
	}
	contents += "\n"

	if options.Print {
		return printMetadata(contents, dir, shouldClean)
	}

	return saveMetadata(options, contents, envExists)
}

func metadataOptions(c *cli.Context) (envset.MetadataOptions, string, bool, error) {
	projectURL, err := gitconfig.OriginURL()
	if err != nil && !isMissingRemoteURL(err) {
		return envset.MetadataOptions{}, "", false, err
	}

	dir, err := filepath.Abs(c.String("filepath"))
	if err != nil {
		return envset.MetadataOptions{}, "", false, err
	}

	shouldClean := false
	if ok := exists(dir); !ok {
		shouldClean = true
		if err = os.MkdirAll(dir, 0750); err != nil {
			return envset.MetadataOptions{}, "", false, err
		}
	}

	algorithm := c.String("hash-algo")
	secret := c.String("secret")
	if secret != "" {
		algorithm = envset.HashHMAC
	}

	return envset.MetadataOptions{
		Name:          cliopts.String(c, cliopts.EnvFileFlag),
		Filepath:      filepath.Join(dir, c.String("filename")),
		Algorithm:     algorithm,
		Project:       projectURL,
		Globals:       c.Bool("globals"),
		GlobalSection: "globals", //TODO: make flag
		Overwrite:     c.Bool("overwrite"),
		Print:         c.Bool("print"),
		Values:        c.Bool("values"),
		Secret:        secret,
	}, dir, shouldClean, nil
}

func metadataChanged(path string, newEnv *envset.EnvFile) (bool, error) {
	oldEnv, err := envset.LoadMetadataFile(path)
	if err != nil {
		return false, err
	}

	return envset.CompareMetadataFiles(newEnv, oldEnv)
}

func printMetadata(contents, dir string, shouldClean bool) error {
	if shouldClean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove metadata dir %s: %w", dir, err)
		}
	}

	if _, err := fmt.Print(contents); err != nil {
		return fmt.Errorf("print output: %w", err)
	}
	return nil
}

func saveMetadata(options envset.MetadataOptions, contents string, envExists bool) error {
	if !envExists {
		if err := writeMetadataFile(options.Filepath, contents); err != nil {
			return fmt.Errorf("write file %s: %w", options.Filepath, err)
		}
		return nil
	}

	if options.Overwrite {
		if err := writeMetadataFile(options.Filepath, contents); err != nil {
			return fmt.Errorf("overwrite file %s: %w", options.Filepath, err)
		}
	}
	return nil
}

func runMetadataCompare(cnf *config.Config, c *cli.Context) error {
	printOutput := c.Bool("print")
	asJSON := c.Bool("json")
	name := c.String("section")
	ignored := cnf.MergeIgnored(name, c.StringSlice("ignore"))

	source, target, err := metadataComparePaths(cnf, c)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if msg := validateMetadataArgs(source, target); msg != "" {
		return cli.Exit(msg, 1)
	}

	s1, s2, err := loadCompareSections(source, target, name)
	if err != nil {
		return err
	}

	diff := envset.CompareSections(*s1, *s2, ignored)
	diff.Name = name

	return reportCompareResult(diff, source, target, ignored, printOutput, asJSON)
}

func metadataComparePaths(cnf *config.Config, c *cli.Context) (string, string, error) {
	if c.Args().Len() != 1 {
		return c.Args().Get(0), c.Args().Get(1), nil
	}

	source, err := envset.FileFinder(filepath.Join(cnf.Meta.Dir, cnf.Meta.File))
	if err != nil {
		return "", "", fmt.Errorf("find default source metadata: %w", err)
	}

	return makeRelative(source), c.Args().Get(0), nil
}

func loadCompareSections(source, target, name string) (*envset.EnvSection, *envset.EnvSection, error) {
	src := envset.EnvFile{}
	if err := src.FromJSON(source); err != nil {
		return nil, nil, cli.Exit(fmt.Sprintf("Unable to load source metadata file %q: %s", source, err), 1)
	}

	s1, err := src.GetSection(name)
	if err != nil {
		fmt.Printf("source: %s\ntarget: %s\nerror: %s\n", source, target, err.Error())
		return nil, nil, cli.Exit(fmt.Sprintf("Section \"%s\" not found in source metadata file:\n%s", name, source), 1)
	}

	tgt := envset.EnvFile{}
	if err := tgt.FromJSON(target); err != nil {
		return nil, nil, cli.Exit(fmt.Sprintf("Unable to load target metadata file %q: %s", target, err), 1)
	}

	s2, err := tgt.GetSection(name)
	if err != nil {
		return nil, nil, cli.Exit(fmt.Sprintf("Section \"%s\" not found in target metadata file.", name), 1)
	}

	return s1, s2, nil
}

func reportCompareResult(diff envset.EnvSection, source, target string, ignored []string, printOutput, asJSON bool) error {
	if diff.IsEmpty() {
		if printOutput && !asJSON {
			prettyOk(source, target)
			return cli.Exit("", 0)
		}
		return nil
	}

	if printOutput && !asJSON {
		prettyPrint(diff, source, target, ignored)
		return cli.Exit("", 1)
	}

	if printOutput && asJSON {
		j, err := diff.ToJSON()
		if err != nil {
			return cli.Exit(err, 1)
		}
		return cli.Exit(j, 1)
	}

	return cli.Exit("Metadata test failed!", 1)
}

func prettyPrint(diff envset.EnvSection, source, target string, ignored []string) {

	sort.Slice(diff.Keys, func(p, q int) bool {
		return diff.Keys[p].Comment > diff.Keys[q].Comment
	})

	fmt.Println("")

	mit := uitable.New()
	mit.MaxColWidth = 50

	mrt := uitable.New()
	mrt.MaxColWidth = 50

	dvt := uitable.New()
	dvt.MaxColWidth = 50

	mi := 0
	mr := 0
	dv := 0

	for _, k := range diff.Keys {
		if strings.Contains(k.Comment, "missing") {
			if mi == 0 {

				mit.AddRow(
					"   "+colors.Bold("STATUS").Underline().String(),
					colors.Bold("ENV KEY").Underline(),
					colors.Bold("HASH").Underline(),
				)
			}
			mi++
			mit.AddRow("👻 Missing", k.Name, strmax(k.Hash, 12, "..."))
		} else if strings.Contains(k.Comment, "extra") {
			if mr == 0 {
				mrt.AddRow(
					"   "+colors.Bold("STATUS").Underline().String(),
					colors.Bold("ENV KEY").Underline(),
					colors.Bold("HASH").Underline(),
				)
			}
			mr++
			mrt.AddRow("🌱 Missing", k.Name, strmax(k.Hash, 12, "..."))
		} else if strings.Contains(k.Comment, "different") {
			if dv == 0 {
				dvt.AddRow(
					"   "+colors.Bold("STATUS").Underline().String(),
					colors.Bold("ENV KEY").Underline(),
					colors.Bold("HASH").Underline(),
				)
			}
			dv++
			dvt.AddRow("❓ Different", k.Name, strmax(k.Hash, 12, "..."))
		}
	}

	fmt.Printf("•  %s: %s\n", colors.Bold("source"), source)
	fmt.Println(tableOrMessage(mit.String(), colors.Green("👍 source is not missing environment variables").String()))

	fmt.Printf("\n\n•  %s: %s\n", colors.Bold("target"), target)
	fmt.Println(tableOrMessage(mrt.String(), colors.Green("👍 target has no extra environment variables").String()))

	fmt.Printf("\n\n•  %s\n", colors.Bold("different values"))
	fmt.Println(tableOrMessage(dvt.String(), colors.Green("👍 All variables have same values").String()))

	//TODO: print ignored keys
	//for _, ik := range diff.Ignored

	fmt.Println("")

	//TODO: add dynamic padding
	fmt.Printf(
		"\n👻 Missing in %s (%d) | 🌱 Missing in %s (%d) \n\n❓ Different values  (%d) | 🤷 Ignored Keys (%d)\n\n",
		colors.Bold("source"),
		greenOrRed(mr).Bold(),
		colors.Bold("target"),
		greenOrRed(mi).Bold(),
		greenOrYellow(dv).Bold(),
		greenOrYellow(len(ignored)).Bold(),
	)
}

func greenOrRed(val int) colors.Value {
	if val == 0 {
		return colors.Green(val)
	}
	return colors.Red(val)
}

func greenOrYellow(val int) colors.Value {
	if val == 0 {
		return colors.Green(val)
	}
	return colors.Yellow(val)
}

func tableOrMessage(tbl, message string) string {
	if tbl == "" {
		return message
	}
	return tbl
}

func strmax(str string, l int, suffix string) string {
	if len(str) <= l {
		return str
	}
	return str[:l] + suffix
}

func prettyOk(source, target string) {
	fmt.Printf("\n•  %s: %s\n", colors.Bold("source"), source)
	fmt.Printf("•  %s: %s\n", colors.Bold("target"), target)
	fmt.Printf("\n🚀 %s\n\n", colors.Bold("All good!").Green())
}

func makeRelative(src string) string {
	path, err := os.Getwd()
	if err != nil {
		return src
	}
	return strings.TrimPrefix(strings.TrimPrefix(src, path), "/")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func isMissingRemoteURL(err error) bool {
	msg := err.Error()
	return msg == "the key remote.origin.url is not found"
}

func writeMetadataFile(path, contents string) error {
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

func validateMetadataArgs(source, target string) string {
	if source == "" {
		return fmt.Sprintf("Source path \"%s\" is not a valid file", source)
	}

	if target == "" {
		return fmt.Sprintf("Source path \"%s\" is not a valid file", source)
	}

	if !exists(source) {
		return fmt.Sprintf("Source path \"%s\" is not a valid file", source)
	}

	if !exists(target) {
		return fmt.Sprintf("Target path \"%s\" is not a valid file", target)
	}

	return ""
}
