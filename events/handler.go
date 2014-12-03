/*
See http://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern for a description of the Publish-Subscribe pattern.
*/
package events

import (
    // "log"
    "fmt"
)
/*
The Publisher in the Publish-Subscribe pattern, or a shorthand for a function which you may call each time you want to inform about a particular event. Publishers can be created by any Topic instance. 
*/
type Publisher func(interface{})
/*
The Subscriber in the Publish-Subscribe pattern, or a shorthand for a function which you may provide, that is invoked whenever an event is published to the Topic. 
*/
type Subscriber func(interface{})
/*
A typical Topic used in a Pub-Sub pattern. The Topic has a name, which in theory should identify it uniquely among other topics. 
The implementation does not use this name, unless for informative reasons. Topics can create Publishers and Subscribers.  
Topics can be closed and this closes the Topic permanently.
*/
type Topic interface {
    //Allows you to create a new Publisher for a Topic. If you provide a callback (optional) here, it is guaranteed to be invoked during the publishing act. Depending on the 
    //implementation it may be also used in a defer() section or called when the the publish failed. This may happen if Publishing is using on a Closed Topic. 
    //Publishing may occur in its own go-routine, and it's not guaranteed that the order you call Publishers is preserved (especially if you write to multiple Topics). 
    NewPublisher() Publisher
    //Allows you to register an arbitrary Subscriber for events in the Topic
    //Subscribing may occur in its own go-routine, hence even if the act of subscribing 'blocks' (for example due to the waiting on channel), the remainders of the Topic still execute normally. 
    NewSubscriber(subscriber Subscriber)
    //Returns the topic's name
    String() string
    //Close frees the underlying resources, and depending on the implementation may render the Topic unusable
    Close() error
}

// ######## Implementations below ###########


type topic struct {
    newSubscribers chan Subscriber
    name string
    events chan interface{}
    finish chan bool
    subscribers []Subscriber
    closed bool
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

func (t *topic) run() {
    go func() {
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(t.subscribers) == 0 {
            t.subscribers = append(t.subscribers, <-t.newSubscribers)
        }
        for ;; {
            if t.closed {
                if t.logger != nil {
                    t.logger(fmt.Sprintf("%v closed.", t))
                }
                return
            }
            select {
            case <-t.finish: //released when you close the channel
                t.closed = true
                return
            case newSubscriber:=<-t.newSubscribers:
                if t.closed {
                    if t.logger != nil {
                        t.logger(fmt.Sprintf("%v closed.", t))
                    }
                    return
                }
                //note: when channel is closed newSubscriber == nil
                if newSubscriber != nil {
                    t.subscribers = append(t.subscribers, newSubscriber)
                }
            case event:=<-t.events:
                if t.closed {
                    if t.logger != nil {
                        t.logger(fmt.Sprintf("%v closed.", t))
                    }
                    return
                }
                //note: when channel is closed event == nil
                if event != nil {
                    if t.logger != nil {
                        t.logger(fmt.Sprintf("%v notifying %v subscribers about %v.", t, len(t.subscribers), event))
                    }
                    for _, subscriber := range t.subscribers {
                        // log.Println(fmt.Sprintf("%v <- %v %T", t.String(), event, event))
                        //note: if subscriber sends something to a channel we don't want to be blocked.
                        go subscriber(event)
                    }
                }
            }
        }
    }()
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
        false,
        nil,
    }
    bus.run()
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
        false,
        loggingMethod,
    }
    bus.run()
    return bus
}
