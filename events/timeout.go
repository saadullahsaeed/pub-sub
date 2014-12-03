package events

import (
    "time"
    "errors"
)

func NewTimeoutingTopic(timeoutTopicName string, timeout time.Duration) Topic {
    return &timeouter {}
}

type timeouter struct {
}

func (t *timeouter) NewPublisher() Publisher {
    return nil
}

func (t *timeouter) NewSubscriber(subscriber Subscriber) <-chan bool {
    return nil
}

func (t *timeouter) String() string {
    return "timeouter"
}

func (t *timeouter) Close() error {
    return nil
}
/**
Allows you to put an artificial timeout on a Topic, and send errors to a designated Topic whenever an event does not arrive in 
a specified amount of time. 

The created topic publishes events which are in fact errors. 
*/
func WhenTimeout(topic Topic, timeout time.Duration, timeoutTopicName string) Topic {
    events := make(chan interface{})
    closeChannel := make(chan bool)
    timeouts := &timeoutingTopic { NewTopic(timeoutTopicName), events, closeChannel }
    publisher := timeouts.NewPublisher()

    subscriber := func(event interface{}) {
        events<-event
    }
    topic.NewSubscriber(subscriber)
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

type timeoutingTopic struct {
    topic Topic
    channel chan interface{}
    closeChannel chan bool
}

func (t *timeoutingTopic) NewPublisher() Publisher {
    return t.topic.NewPublisher()
}

func (t *timeoutingTopic) NewSubscriber(subscriber Subscriber) <-chan bool {
    return t.topic.NewSubscriber(subscriber)
}

func (t *timeoutingTopic) String() string {
    return t.topic.String()
}

func (t *timeoutingTopic) Close() error {
    close(t.channel)
    close(t.closeChannel)
    return t.topic.Close()
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
