package version

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// TODO: This package should be build rather han version, so we can Tag -> Version
var (
	Tag    = "dev"
	Time   string
	User   string
	Commit string
)

// GetVersion returns version string
func GetVersion() string {
	return Tag + "-" + Time + ":" + User
}

// Print will output our version in a format
func Print(w io.Writer) error {
	tw := new(tabwriter.Writer)
	tw.Init(w, 0, 0, 0, ' ', tabwriter.AlignRight)
	if _, err := fmt.Fprintln(tw); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, "Version:", "\t", Tag); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, "Build Commit Hash:", "\t", Commit); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, "Build Time:", "\t", Time); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, "Build User:", "\t", User); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw); err != nil {
		return err
	}
	return tw.Flush()
}
