package version

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
