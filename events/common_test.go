package events

import (
    "log"
    "fmt"
)

func defaultLogging(message string, args ...interface{}) {
    log.Println(fmt.Sprintf(message, args))
}
