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
    closeChannel := make(chan bool)
    subscriber := func(event interface{}) {
        events<-event
    }
    topic.NewSubscriber(subscriber)
    timeouts := &topicWithChannel { NewTopic(timeoutTopicName), events, closeChannel }
    publisher := timeouts.NewPublisher(nil)
    andListen := func() {
        var (
            timeoutChan <-chan time.Time
        )
        timeoutChan = time.After(timeout)
        for ;; {
            select {
            case <-closeChannel:
                return
            case <-events:
                timeoutChan = time.After(timeout)
            case <-timeoutChan:
                publisher(errors.New("Timeout on "+topic.String()))
            }
        }
    }
    go andListen()
    return timeouts
}

/**
This a variant of WhenTimeout, which panics instead of sending errors on a Topic
*/
func MustPublishWithin(topic Topic, timeout time.Duration) {
    errorTopic := WhenTimeout(topic, timeout, topic.String()+"-timeout-errors-collector")
    errorTopic.NewSubscriber(func(err interface{}) {
        go errorTopic.Close()
        panic(err.(error))
    })
}

type topicWithChannel struct {
    topic Topic
    channel chan interface{}
    closeChannel chan bool
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
    close(t.closeChannel)
    return t.topic.Close()
}

