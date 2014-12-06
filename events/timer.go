package events

import (
    "time"
    "fmt"
    "errors"
)

/**
Produces a Topic which encapsulates a Go <-time.NewTicker(). 

Ticker topics have one important distinction over a classic 'Topic': publishing does not actually cause 
a Subscribe event on the other end. Publishing to the Topic causes a ticker reset, but the event itself 
is otherwise ignored. 
*/
func NewTickerTopic(topicName string, timeout time.Duration) Topic {
    bus := &ticker {
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
    andRunLoop := buildTickerLoop(bus.spec, bus.timeout)
    go andRunLoop()
    return bus
}

func NewTickerTopicWithLogging(topicName string, timeout time.Duration, loggingMethod func(...interface{})) Topic {
    bus := &ticker {
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
    andRunLoop := buildTickerLoop(bus.spec, bus.timeout)
    go andRunLoop()
    return bus
}

type ticker struct {
    spec *topicSpec
    timeout time.Duration
}

func (t *ticker) NewPublisher() Publisher {
    return t.spec.NewPublisher()
}

func (t *ticker) NewSubscriber(subscriber Subscriber) <-chan bool {
    return t.spec.NewSubscriber(subscriber)
}

func (t *ticker) String() string {
    return t.spec.String()
}

func (t *ticker) Close() error {
    return t.spec.Close()
}
/**
Allows you to put an artificial timeout on a Topic, and send time.Time events to a designated Topic whenever an event does not arrive in 
a specified amount of time. 
*/
func WhenTimeout(topic Topic, timeout time.Duration, tickerTopicName string) Topic {
    return whenTimeout(topic, NewTickerTopic(tickerTopicName, timeout))
}

func WhenTimeoutWithLogging(topic Topic, timeout time.Duration, tickerTopicName string, loggingMethod func(...interface{})) Topic {
    return whenTimeout(topic, NewTickerTopicWithLogging(tickerTopicName, timeout, loggingMethod))
}

func whenTimeout(topic Topic, tickerTopic Topic) Topic {
    tickerPublisher := tickerTopic.NewPublisher()
    subscriber := func(event interface{}) {
        tickerPublisher(event)
    }
    topic.NewSubscriber(subscriber)
    return tickerTopic
}

/**
This a variant of WhenTimeout, which panics when something is received in the underlying Ticker topic. 
*/
func MustPublishWithin(topic Topic, timeout time.Duration) {
    errorTopic := WhenTimeout(topic, timeout, topic.String()+"-timeout-errors-collector")
    errorTopic.NewSubscriber(func(timeout interface{}) {
        go errorTopic.Close()
        panic(errors.New(fmt.Sprintf("%v: Timeout at %v", topic, timeout)))
    })
}
