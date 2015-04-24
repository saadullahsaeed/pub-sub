package events

import (
	"io/ioutil"
	"fmt"
)
const (
	buildNumber = ".minor.version"
)

//Returns the current version of the library
func Version() string {
	version, err := ioutil.ReadFile(buildNumber)
	if err != nil {
		return "no tag present"
	}
	return fmt.Sprintf("2.1.%v", string(version))
}
