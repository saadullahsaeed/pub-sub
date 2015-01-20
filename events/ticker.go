package events

import (
    "time"
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
            make(chan string),
            []Subscriber{},
            nil,
        },
        timeout,
    }
    andRunLoop := buildTickerLoop(bus.spec, bus.timeout)
    go andRunLoop()
    return bus
}

func NewTickerTopicWithLogging(topicName string, timeout time.Duration, loggingMethod func(string, ...interface{})) Topic {
    bus := &ticker {
            &topicSpec {
                make(chan Subscriber),
                topicName,
                make(chan interface{}),
                make(chan string),
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
