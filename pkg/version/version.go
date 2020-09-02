package version

var (
	BuildVersion string = ""
	BuildTime    string = ""
	BuildUser    string = ""
)

//GetVersion returns version string
func GetVersion() string {
	return BuildVersion + "-" + BuildTime + ":" + BuildUser
}
