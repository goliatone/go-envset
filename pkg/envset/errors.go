package envset

import "fmt"

type envFileErrorNotFound struct {
	err error
	msg string
}

func (e envFileErrorNotFound) Error() string {
	return e.msg
}

//IsFileNotFound will return true if v is file not found
func IsFileNotFound(v interface{}) bool {
	_, isType := v.(envFileErrorNotFound)
	return isType
}

type envSectionErrorNotFound struct {
	err error
	msg string
}

func (e envSectionErrorNotFound) Error() string {
	return e.msg
}

//IsSectionNotFound will return true if v is section not found
func IsSectionNotFound(v interface{}) bool {
	_, isType := v.(envSectionErrorNotFound)
	return isType
}

type ErrorRunningCommand struct {
	err error
	msg string
}

func (e ErrorRunningCommand) Error() string {
	return e.msg
}

//IsErrorRunningCommand will return true if v is section not found
func IsErrorRunningCommand(v interface{}) bool {
	_, isType := v.(ErrorRunningCommand)
	return isType
}

//ErrorWrongAlgorithm generated when source and target have different
//algorithms
type ErrorWrongAlgorithm struct {
	source string
	target string
}

func (e *ErrorWrongAlgorithm) Error() string {
	return fmt.Sprintf("wrong algorithm: source %s target %s", e.source, e.target)
}
