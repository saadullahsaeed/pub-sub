/*
See http://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern for a description of the Publish-Subscribe pattern.
*/
package events

import (
    "io"
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
Topics can be closed -- this closes the Topic permanently.
*/
type Topic interface {
    //Allows you to create a new Publisher for a Topic. If you provide a callback (optional) here, it is guaranteed to be invoked during the publishing act.
    //Publishing may occur in its own go-routine, and it's not guaranteed that the order you call Publishers is preserved (especially if you write to multiple Topics). 
    NewPublisher(optionalCallback Publisher) Publisher
    //Allows you to register an arbitrary Subscriber for events in the Topic
    //Subscribing may occur in its own go-routine, hence even if the act of subscribing 'blocks' (for example due to the waiting on channel), the remainders of the Topic still execute normally. 
    NewSubscriber(subscriber Subscriber)
    //Returns the topic's name
    String() string
    io.Closer
}

// ######## Implementations below ###########


type topic struct {
    newSubscribers chan Subscriber
    name string
    events chan interface{}
    finish chan bool
    subscribers []Subscriber
}

func (t *topic) NewPublisher(optionalCallback Publisher) Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
            t.events<- event
            //client will not know something is wrong (deadlock?) if they provide a dodgy callback
            if optionalCallback != nil {
                optionalCallback(event)
            }
        }()
    }
    return publisher
}
func (t *topic) NewSubscriber(subscriber Subscriber) {
    t.newSubscribers<-subscriber
}
func (t *topic) String() string {
    return t.name
}
func (t *topic) Close() error {
    close(t.newSubscribers)
    close(t.events)
    close(t.finish)
    return nil
}

func (t *topic) run() {
    go func() {
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(t.subscribers) == 0 {
            t.subscribers = append(t.subscribers, <-t.newSubscribers)
        }
        for ;; {
            select {
            case <-t.finish:
                return
            case newSubscriber:=<-t.newSubscribers:
                //note: when channel is closed newSubscriber == nil
                if newSubscriber != nil {
                    t.subscribers = append(t.subscribers, newSubscriber)
                }
            case event:=<-t.events:
                //note: when channel is closed event == nil
                if event != nil {
                    for _, subscriber := range t.subscribers {
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
The optionalCallback (passed as a parameter to the Publisher) is invoked (in this implementation) after the publishing actually occured. 
One of the tests to this library, shows an example how to 'synchronise' Publisher execution with the use of a channel: this adds overhead, though.
*/
func NewTopic(topicName string) Topic {
    bus := &topic { make(chan Subscriber), topicName, make(chan interface{}), make(chan bool), []Subscriber{} }
    bus.run()
    return bus
}

