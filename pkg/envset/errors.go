package envset

type envFileErrorNotFound struct {
	err error
	msg string
}

func (e envFileErrorNotFound) Error() string {
	return e.msg
}

type envSectionErrorNotFound struct {
	err error
	msg string
}

func (e envSectionErrorNotFound) Error() string {
	return e.msg
}

//IsFileNotFound will return true if v is file not found
func IsFileNotFound(v interface{}) bool {
	_, isType := v.(envFileErrorNotFound)
	return isType
}

//IsSectionNotFound will return true if v is section not found
func IsSectionNotFound(v interface{}) bool {
	_, isType := v.(envSectionErrorNotFound)
	return isType
}
