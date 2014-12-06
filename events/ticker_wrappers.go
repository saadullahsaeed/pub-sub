package events

import (
    "time"
    "fmt"
    "errors"
)

const (
    TICKER_SUFFIX = "timeouts-ticker"
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

func generateTickerName(topic Topic) string {
    return fmt.Sprintf("%v-%v-%v", topic.String(), TICKER_SUFFIX, time.Now().Unix())
}

/** 
This function awaits for an event on provided Topic for the defined amount of time, and if none occurs - it panics. 
*/
func MustPublishAtLeastOnce(topic Topic, timeout time.Duration) {
    tickerTopic := NewTickerTopic(generateTickerName(topic), timeout)
    mustPublishAtLeastOnce(topic, tickerTopic, timeout)
}

/** 
This is a variant of the MustPublishAtLeastOnce method which allows for logging.
*/
func MustPublishAtLeastOnceWithLogging(topic Topic, timeout time.Duration, loggingMethod func(...interface{})) {
    tickerTopic := NewTickerTopicWithLogging(generateTickerName(topic), timeout, loggingMethod)
    mustPublishAtLeastOnce(topic, tickerTopic, timeout)
}

func mustPublishAtLeastOnce(topic, tickerTopic Topic, timeout time.Duration) {
    joint := Or([]Topic { topic, tickerTopic }, tickerTopic.String()+"-"+topic.String())
    jointSubscriber := func(event interface{}) {
        jointEvent := event.(map[string][]interface{})
        originalEvent, originalEventExists := jointEvent[topic.String()]
        if !originalEventExists || len(originalEvent) == 0 {
            tickerTopic.Close()
            joint.Close()
            panic(errors.New(fmt.Sprintf("%v: No event occured in topic in defined time", topic)))
        }
        joint.Close()
        tickerTopic.Close()
    }
    <-joint.NewSubscriber(jointSubscriber)
}
