package version

import (
	"fmt"
	"io"
	"text/tabwriter"
)

//TODO: This package should be build rather han version, so we can Tag -> Version
var (
	Tag    = "dev"
	Time   string
	User   string
	Commit string
)

//GetVersion returns version string
func GetVersion() string {
	return Tag + "-" + Time + ":" + User
}

//Print will output our version in a format
func Print(w io.Writer) error {
	tw := new(tabwriter.Writer)
	tw.Init(w, 0, 0, 0, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "Version:", "\t", Tag)
	fmt.Fprintln(tw, "Build Commit Hash:", "\t", Commit)
	fmt.Fprintln(tw, "Build Time:", "\t", Time)
	fmt.Fprintln(tw, "Build User:", "\t", User)
	fmt.Fprintln(tw)
	return tw.Flush()
}
