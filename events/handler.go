package events

import (
    // "log"
    "fmt"
)

type topic struct {
    newSubscribers chan Subscriber
    name string
    events chan interface{}
    finish chan bool
    subscribers []Subscriber
    logger func(...interface{})
}

func (t *topic) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
            defer func() {
                if err := recover(); err != nil {
                    if t.logger != nil {
                        t.logger(fmt.Sprintf("%v is most probably closed. Error whilst running NewPublisher(): %v", t, err))
                    }
                    panic(err.(error).Error())
                }
            }()
            t.events<- event
        }()
    }
    return publisher
}
func (t *topic) NewSubscriber(subscriber Subscriber) {
    go func() {
        defer func() {
            if err := recover(); err != nil {
                if t.logger != nil {
                    t.logger(fmt.Sprintf("%v is most probably closed. Error whilst running NewSubscriber(): %v", t, err))
                }
                panic(err.(error).Error())
            }
        }()
        t.newSubscribers<-subscriber
    }()
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
        make(chan Subscriber),
        topicName,
        make(chan interface{}),
        make(chan bool),
        []Subscriber{},
        nil,
    }
    <-runTopicGoRoutine(bus.newSubscribers, bus.name, bus.events, bus.finish, bus.subscribers, nil)
    return bus
}

/* 
This very similar to the NewTopic function, except that it allows for logging (log.Println) to be injected into the framework, which 
is useful for tests.
*/
func NewTopicWithLogging(topicName string, loggingMethod func(...interface{})) Topic {
    bus := &topic {
        make(chan Subscriber),
        topicName,
        make(chan interface{}),
        make(chan bool),
        []Subscriber{},
        loggingMethod,
    }
    <-runTopicGoRoutine(bus.newSubscribers, bus.name, bus.events, bus.finish, bus.subscribers, nil)
    return bus
}
