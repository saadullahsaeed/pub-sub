package events

import (
    "time"
    "errors"
)

/**
Allows you to put an artificial timeout on a Topic, and send errors to a designated Topic whenever an event does not arrive in 
a specified amount of time. 

The created topic publishes events which are in fact errors. 
*/
func WhenTimeout(topic Topic, timeout time.Duration, timeoutTopicName string) Topic {
    events := make(chan interface{})
    subscriber := func(event interface{}) {
        events<-event
    }
    topic.NewSubscriber(subscriber)
    timeouts := &topicWithChannel { NewTopic(timeoutTopicName), events }
    publisher := timeouts.NewPublisher(nil)
    andListen := func() {
        for ;; {
            select {
                case <-events:
                    //ignore
                case <-time.After(timeout):
                    publisher(errors.New("Timeout on "+topic.String()))
            }
        }
    }
    go andListen()
    return timeouts
}

type topicWithChannel struct {
    topic Topic
    channel chan interface{}
}

func (t *topicWithChannel) NewPublisher(optionalCallback func(interface{})) Publisher {
    return t.topic.NewPublisher(optionalCallback)
}

func (t *topicWithChannel) NewSubscriber(subscriber Subscriber) {
    t.topic.NewSubscriber(subscriber)
}

func (t *topicWithChannel) String() string {
    return t.topic.String()
}

func (t *topicWithChannel) Close() error {
    close(t.channel)
    return t.topic.Close()
}

