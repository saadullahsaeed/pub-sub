package events

import (
    "fmt"
)

type topicSpec struct {
    newSubscribers chan Subscriber
    name string
    events chan interface{}
    finish chan bool
    subscribers []Subscriber
    loggingMethod func(...interface{})
}

func (spec *topicSpec) String() string {
    return spec.name
}

func optionallyLogPanics(spec *topicSpec, message string) {
    if err := recover(); err != nil {
        if spec.loggingMethod != nil {
            spec.loggingMethod(fmt.Sprintf("%v: "+message, spec))
        }
        panic(err.(error).Error())
    }
}

func optionallyLog(spec *topicSpec, message string) {
    if spec.loggingMethod != nil {
        spec.loggingMethod(fmt.Sprintf("%v: "+message, spec))
    }
}
