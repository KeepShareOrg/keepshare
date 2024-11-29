package server

import (
	"strings"
)

func IsFileNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "file_not_found")
}

func IsForbiddenShareResourceError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Malicious or extremist content resources, sharing is not supported")
}
