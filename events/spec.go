package events

import (
    // "fmt"
)

type topicSpec struct {
    newSubscribers chan Subscriber
    name string
    events chan interface{}
    finish chan bool
    subscribers []Subscriber
    loggingMethod func(string,...interface{})
}

func (spec *topicSpec) String() string {
    return spec.name
}

func (spec *topicSpec) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
            defer optionallyLogPanics(spec, "probably closed, error in NewPublisher()")
            spec.events<- event
            optionallyLog(spec, "new event")
        }()
    }
    return publisher
}

func (spec *topicSpec) NewSubscriber(subscriber Subscriber) <-chan bool {
    releaser := make(chan bool)
    go func() {
        defer optionallyLogPanics(spec, "probably closed, error in NewSubscriber()")
        spec.newSubscribers<-subscriber
        close(releaser) //this releases awaiting listeners
    }()
    return releaser
}

func (spec *topicSpec) Close() error {
    close(spec.finish)
    close(spec.newSubscribers)
    close(spec.events)
    return nil
}

func optionallyLogPanics(spec *topicSpec, message string) {
    if err := recover(); err != nil {
        if spec.loggingMethod != nil {
            spec.loggingMethod("%v: "+message+"%v", spec, "")
        }
        panic(err.(error).Error())
    }
}

func optionallyLog(spec *topicSpec, message string) {
    if spec.loggingMethod != nil {
        spec.loggingMethod("%v: "+message+"%v", spec, "")
    }
}
