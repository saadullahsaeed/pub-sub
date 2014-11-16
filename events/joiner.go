package events

import (
    "time"
    "fmt"
    "errors"
)

func And(inputTopics []Topic, name string) Topic {
    collectedEventsChannel := make(chan map[string]interface{})
    for _, topic := range inputTopics {
        topic.NewSubscriber(whichCollectsToACommonChannel(collectedEventsChannel, topic.String()))
    }
    outputTopic := &topicWithAChannel{ NewTopic(name), collectedEventsChannel }
    outputTopicPublisher := outputTopic.NewPublisher()
    go andListenOnCommonChannel(collectedEventsChannel, outputTopicPublisher, len(inputTopics))
    return outputTopic
}

func Or(inputTopics []Topic, name string) Topic {
    collectedEventsChannel := make(chan map[string]interface{})
    for _, topic := range inputTopics {
        topic.NewSubscriber(whichCollectsToACommonChannel(collectedEventsChannel, topic.String()))
    }
    outputTopic := &topicWithAChannel{ NewTopic(name), collectedEventsChannel }
    outputTopicPublisher := outputTopic.NewPublisher()
    go andListenOnCommonChannel(collectedEventsChannel, outputTopicPublisher, 1)
    return outputTopic
}

func whichCollectsToACommonChannel(namedEventChannel chan map[string]interface{}, topicName string) func(interface{}) {
    return func(input interface{}) {
        defer func() {
            if err := recover(); err != nil {
                //usually we want to ignore this. 
                //panics occur here, when the main Topic is closed, and there is still something sent by a client.
                //it's actually a client error, as from the point of view of our code, the Topic has been closed 
                //before the Publish...
            }
        }()
        namedEventChannel<- map[string]interface{} {
            topicName : input,
        }
    }
}

func andListenOnCommonChannel(namedEventChannel chan map[string]interface{}, publisher func(interface{}), expectedSize int) {
    collectedResults := map[string][]interface{} {}
    for ;; {
        newEvent := <-namedEventChannel
        for topicName, anEventInATopic := range newEvent {
            collectedResults[topicName] = append(collectedResults[topicName], anEventInATopic)
        }
        currentSize := 0
        for _, array := range collectedResults {
            if len(array) > 0 {
                currentSize = currentSize + 1
            }
        }
        if currentSize == expectedSize {
            publisher(copyAside(collectedResults))
            for topicName, _ := range collectedResults {
                delete(collectedResults,topicName)
            }
        }
    }
}

func copyAside(original map[string][]interface{}) map[string][]interface{} {
    mapCopy := map[string][]interface{} {}
    for key,value := range original {
        mapCopy[key] = value
    }
    return mapCopy
}

type topicWithAChannel struct {
    topic Topic
    channel chan map[string]interface{}
}

func (t *topicWithAChannel) NewPublisher() Publisher {
    return t.topic.NewPublisher()
}

func (t *topicWithAChannel) NewSubscriber(subscriber Subscriber) {
    t.topic.NewSubscriber(subscriber)
}

func (t *topicWithAChannel) String() string {
    return t.topic.String()
}

func (t *topicWithAChannel) Close() error {
    close(t.channel)
    return t.topic.Close()
}

