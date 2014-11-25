package events

import (
    "time"
    "log"
    "fmt"
)

/**
Creates a wrapper over an existing Topic (and delegates all calls to it) and a go-routine which closes the topic after a
provided timeout. This is useful if you do not want to control the longevity of a Topic
*/
func NewTemporaryTopic(topic Topic, timeout time.Duration) Topic {
    temporaryTopic := &temporaryTopic { topic }
    go andCloseAfterTimeout(temporaryTopic, timeout)
    return temporaryTopic
}

type temporaryTopic struct {
    topic Topic
}

func (t *temporaryTopic) NewPublisher(optionalCallback func(interface{})) Publisher {
    defer recoveryFromPanics()
    return t.topic.NewPublisher(optionalCallback)
}

func (t *temporaryTopic) NewSubscriber(subscriber Subscriber) {
    defer recoveryFromPanics()
    t.topic.NewSubscriber(subscriber)
}

func (t *temporaryTopic) String() string {
    return t.topic.String()
}

func (t *temporaryTopic) Close() error {
    defer recoveryFromPanics()
    return t.topic.Close()
}

func andCloseAfterTimeout(topic Topic, timeout time.Duration) {
    <-time.After(timeout)
    topic.Close()
}

func recoveryFromPanics() {
    if err := recover(); err!=nil {
        log.Println(fmt.Sprintf("Can't publish on temporary topic, perhaps because it's closed? %v", err))
    }
}


