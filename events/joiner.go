package events

import (
    // "time"
    // "fmt"
    // "errors"
    "log"
)

type collectedResults map[string][]interface{}
type result map[string]interface{}

/**
Allows you to join several topics together into a single one, with an AND operator. This means, the output topic fires off an event whenever all
provided Topics have had Publish events. 

The resulting Topic publishes events which are actually map[string][]interface{} structures: the key points to the originating Topic, whilst the value is an 
array of all events captured so far within the Topic, in the order they have been received (note: not necessarily the order in which they have been sent).
Do note, that the same kind of structure is passed as an event, when using the alternative to this function, called OR: as a conseuquence, you may use
the same event handling code in both situations, and chain these topics together into longer pipes.
*/
func And(inputTopics []Topic, name string) Topic {
    return awaitEvent(inputTopics, name, len(inputTopics))
}

/**
Allows you to join several topics together into a single one, with an OR operator. This means, the output topic fires off an event whenever any of the 
provided Topics have had Publish events. 

The resulting Topic publishes events which are actually map[string][]interface{} structures: the key points to the originating Topic, whilst the value is an 
array of all events captured so far within the Topic, in the order they have been received (note: not necessarily the order in which they have been sent).
Contrary, to the AND function, OR operator is simpler: since events in the resulting Topic are fired whenever any of the incoming Topics has an event, 
the returned structure could be simplified. Still, the writer of this library believes that having a uniform structure behind AND and OR is a higher value, 
than the potential benefit gained by reducing the complexity here. As a consequence, handling code can be exchanged between AND and OR, and chained together
into longer pipelines.
*/
func Or(inputTopics []Topic, name string) Topic {
    return awaitEvent(inputTopics, name, 1)
}

func awaitEvent(inputTopics []Topic, name string, releaseResultsWhenSizeReached int) Topic {
    newStates := make(chan collectedResults)
    currentState := make(chan collectedResults)
    for _, topic := range inputTopics {
        topic.NewSubscriber(whichCollectsToACommonChannel(newStates, currentState, topic.String()))
    }
    outputTopic := &topicWithAChannel{ NewTopic(name), newStates, currentState }
    outputTopicPublisher := outputTopic.NewPublisher(nil)
    go andListen(newStates, currentState, outputTopicPublisher, releaseResultsWhenSizeReached)
    newStates<-map[string][]interface{} {}
    return outputTopic
}

func whichCollectsToACommonChannel(newStates, currentState chan collectedResults, topicName string) Subscriber {
    return func(input interface{}) {
        defer func() {
            if err := recover(); err != nil {
                //usually we want to ignore this. 
                //panics occur here, when the main Topic is closed, and there is still something sent by a client.
                //it's actually a client error, as from the point of view of our code, the Topic has been closed 
                //before the Publish...
            }
        }()
        state := <-currentState
        state[topicName] = append(state[topicName], input)
        newStates<-state
    }
}

func andListen(newStates, currentState chan collectedResults, publisher func(interface{}), releaseResultsWhenSizeReached int) {
    for ;; {
        newState := <-newStates
        currentSize := 0
        for _, array := range newState {
            if len(array) > 0 {
                currentSize = currentSize + 1
            }
        }
        log.Println(newState)
        if currentSize == releaseResultsWhenSizeReached {
            log.Println("Bong")
            publisher(copyAside(newState))
            for topicName, _ := range newState {
                delete(newState,topicName)
            }
        }
        currentState<-newState
    }
}

func copyAside(original collectedResults) collectedResults {
    mapCopy := collectedResults {}
    for key,value := range original {
        copiedValue := []interface{} {}
        for _, arrayValue := range value {
            copiedValue = append(copiedValue, arrayValue)
        }
        mapCopy[key] = copiedValue
    }
    return mapCopy
}

type topicWithAChannel struct {
    topic Topic
    in chan collectedResults
    out chan collectedResults
}

func (t *topicWithAChannel) NewPublisher(optionalCallback Publisher) Publisher {
    return t.topic.NewPublisher(optionalCallback)
}

func (t *topicWithAChannel) NewSubscriber(subscriber Subscriber) {
    t.topic.NewSubscriber(subscriber)
}

func (t *topicWithAChannel) String() string {
    return t.topic.String()
}

func (t *topicWithAChannel) Close() error {
    close(t.in)
    close(t.out)
    return t.topic.Close()
}

