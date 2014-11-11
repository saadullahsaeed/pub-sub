/*
See http://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern for a description of the Publish-Subscribe pattern.
*/
package handlers

/*
The Publisher in the Publish-Subscribe pattern, or a shorthand for a function which you may call each time you want to inform about a particular event. 
Publishers can be created by any Topic instance. 
*/
type Publisher func(interface{})
/*
The Subscriber in the Publish-Subscribe pattern, or a shorthand for a function which you may provide, that is invoked whenever an event is published to the Topic. 
*/
type Subscriber func(interface{})
/*
A typical Topic used in a Pub-Sub pattern. The Topic has a name, which in theory should identify it uniquely among other topics. 
The implementation does not use this name, unless for informative reasons. Topics can create Publishers and Subscribers.  
*/
type Topic interface {
    //Allows you to create a new Publisher for a Topic. 
    NewPublisher() Publisher
    //Allows you to register an arbitrary Subscriber for events in the Topic
    NewSubscriber(subscriber Subscriber)
    //Returns the topic's name
    String() string
}

// ######## Implementations below ###########


type topic struct {
    newSubscribers chan Subscriber
    name string
    events chan interface{}
    subscribers []Subscriber
}

func (t *topic) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        t.events<- event
    }
    return publisher
}
func (t *topic) NewSubscriber(subscriber Subscriber) {
    go func() {
        t.newSubscribers<-subscriber
    }()
}
func (t *topic) String() string {
    return t.name
}

func (t *topic) run() {
    go func() {
        for ;; {
            select {
            case newSubscriber:=<-t.newSubscribers:
                t.subscribers = append(t.subscribers, newSubscriber)
            case event:=<-t.events:
                for _, subscriber := range t.subscribers {
                    subscriber(event)
                }
            }
        }
    }()
}

/* 
Creates a new Topic that can create Publishers and which Subscribers can attach to. 
The 'buffer' parameter specifies the size of an internal channel that determines when 'blocking' starts to occur. If you don't know what to put there, specify a 100 or so. Otherwise, consult 
https://golang.org/doc/effective_go.html#channels. 
*/
func NewTopic(topicName string) Topic {
    bus := &topic { make(chan Subscriber), topicName, make(chan interface{}), []Subscriber{} }
    bus.run()
    return bus
}


