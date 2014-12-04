package events

import (
    "time"
    // "fmt"
    // "errors"
)

/**
Produces a Topic which encapsulates a Go <-time.After() timer. 

Publishing to the Topic causes a Timer reset, but the event itself is otherwise ignored.

*/
func NewTimerTopic(topicName string, timeout time.Duration) Topic {
    bus := &timer {
        &topicSpec {
            make(chan Subscriber),
            topicName,
            make(chan interface{}),
            make(chan bool),
            []Subscriber{},
            nil,
        },
        timeout,
    }
    andRunLoop := buildTimerLoop(bus.spec, bus.timeout)
    go andRunLoop()
    return bus
}

func NewTimerTopicWithLogging(topicName string, timeout time.Duration, loggingMethod func(...interface{})) Topic {
    bus := &timer {
            &topicSpec {
                make(chan Subscriber),
                topicName,
                make(chan interface{}),
                make(chan bool),
                []Subscriber{},
                loggingMethod,
        },
        timeout,
    }
    andRunLoop := buildTimerLoop(bus.spec, bus.timeout)
    go andRunLoop()
    return bus
}

type timer struct {
    spec *topicSpec
    timeout time.Duration
}

func (t *timer) NewPublisher() Publisher {
    return t.spec.NewPublisher()
}

func (t *timer) NewSubscriber(subscriber Subscriber) <-chan bool {
    return t.spec.NewSubscriber(subscriber)
}

func (t *timer) String() string {
    return t.spec.String()
}

func (t *timer) Close() error {
    return t.spec.Close()
}
/**
Allows you to put an artificial timeout on a Topic, and send errors to a designated Topic whenever an event does not arrive in 
a specified amount of time. 

The created topic publishes events which are in fact errors. 
*/
func WhenTimeout(topic Topic, timeout time.Duration, timeoutTopicName string) Topic {
    timerTopic := NewTimerTopic(timeoutTopicName, timeout)
    timingPublisher := timerTopic.NewPublisher()
    subscriber := func(event interface{}) {
        timingPublisher(event)
    }
    topic.NewSubscriber(subscriber)
    return timerTopic
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
