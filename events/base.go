package events

import (
    // "log"
    // "fmt"
)

/* 
Creates a new Topic that can create Publishers and Subscribers that run in separate go-routines. This means order of Publishing may not be guaranteed to some extent. 
Likewise, the order in which Subscribers are called can't be fully guaranteed.
*/
func NewTopic(topicName string) Topic {
    bus := &topicSpec {
            make(chan Subscriber),
            topicName,
            make(chan interface{}),
            make(chan string),
            []Subscriber{},
            nil,
    }
    runner := buildBaseLoop(bus)
    go runner()
    return bus
}

/* 
This very similar to the NewTopic function, except that it allows for logging (log.Println) to be injected into the framework, which 
is useful for tests.
*/
func NewTopicWithLogging(topicName string, loggingMethod func(string,...interface{})) Topic {
    bus := &topicSpec {
            make(chan Subscriber),
            topicName,
            make(chan interface{}),
            make(chan string),
            []Subscriber{},
            loggingMethod,
    }
    runner := buildBaseLoop(bus)
    go runner()
    return bus
}



