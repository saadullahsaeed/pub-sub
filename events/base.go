package events

import (
    // "log"
    // "fmt"
)

type topic struct {
    topicSpec
}

func (t *topic) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
            defer optionallyLogPanics(&t.topicSpec, "topic probably closed, error in NewPublisher()")
            t.events<- event
        }()
    }
    return publisher
}

func (t *topic) NewSubscriber(subscriber Subscriber) <-chan bool {
    releaser := make(chan bool)
    go func() {
        defer optionallyLogPanics(&t.topicSpec, "topic probably closed, error in NewSubscriber()")
        t.newSubscribers<-subscriber
        close(releaser) //this releases awaiting listeners
    }()
    return releaser
}

func (t *topic) String() string {
    return t.name
}

func (t *topic) Close() error {
    close(t.finish)
    close(t.newSubscribers)
    close(t.events)
    return nil
}

/* 
Creates a new Topic that can create Publishers and Subscribers that run in separate go-routines. This means order of Publishing may not be guaranteed to some extent. 
Likewise, the order in which Subscribers are called can't be fully guaranteed.
*/
func NewTopic(topicName string) Topic {
    bus := &topic {
        topicSpec {
            make(chan Subscriber),
            topicName,
            make(chan interface{}),
            make(chan bool),
            []Subscriber{},
            nil,
        },
    }
    runner := buildBaseLoop(&(bus.topicSpec))
    go runner()
    return bus
}

/* 
This very similar to the NewTopic function, except that it allows for logging (log.Println) to be injected into the framework, which 
is useful for tests.
*/
func NewTopicWithLogging(topicName string, loggingMethod func(...interface{})) Topic {
    bus := &topic {
        topicSpec {
            make(chan Subscriber),
            topicName,
            make(chan interface{}),
            make(chan bool),
            []Subscriber{},
            loggingMethod,
        },
    }
    runner := buildBaseLoop(&(bus.topicSpec))
    go runner()
    return bus
}
