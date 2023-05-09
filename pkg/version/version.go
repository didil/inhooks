package version

var ver string

func SetVersion(v string) {
	ver = v
}

func GetVersion() string {
	return ver
}
