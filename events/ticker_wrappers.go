package events

import (
    "time"
    "fmt"
    "errors"
)
/**
Allows you to put an artificial timeout on a Topic, and send time.Time events to a designated Topic whenever an event does not arrive in 
a specified amount of time. 
*/
func WhenTimeout(topic Topic, timeout time.Duration, tickerTopicName string) Topic {
    return whenTimeout(topic, NewTickerTopic(tickerTopicName, timeout))
}

/**
A variant of WhenTimeout that allows for logging.
*/
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
func MustPublish(topic Topic, timeout time.Duration) {
    errorTopic := WhenTimeout(topic, timeout, topic.String()+"-timeout-errors-collector")
    errorTopic.NewSubscriber(func(timeout interface{}) {
        go errorTopic.Close()
        panic(errors.New(fmt.Sprintf("%v: Timeout at %v", topic, timeout)))
    })
}

/**
This a variant of MustPublish that allows for logging.
*/
func MustPublishWithLogging(topic Topic, timeout time.Duration, loggingMethod func(...interface{})) {
    errorTopic := WhenTimeoutWithLogging(topic, timeout, topic.String()+"-timeout-errors-collector", loggingMethod)
    errorTopic.NewSubscriber(func(timeout interface{}) {
        go errorTopic.Close()
        panic(errors.New(fmt.Sprintf("%v: Timeout at %v", topic, timeout)))
    })
}
