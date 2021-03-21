//TODO: This package should be build rather han version, so we can Tag -> Version
package version

var (
	Tag 	string = ""
	Time    string = ""
	User    string = ""
	Commit  string = ""
)

//GetVersion returns version string
func GetVersion() string {
	return Tag + "-" + Time + ":" + User
}
